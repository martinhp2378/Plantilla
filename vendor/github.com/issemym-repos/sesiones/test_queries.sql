drop schema if exists sesion cascade;
drop schema if exists institucion cascade;
create schema institucion;
create schema sesion;

create table institucion.clinica
(
    uuid_clinica uuid primary key         default gen_random_uuid(),
    nombre       varchar(250) not null
);
insert into institucion.clinica
(nombre)
values
('Clinica 1'),
('Clinica 2');

create table institucion.persona
(
    uuid_persona  uuid primary key         default gen_random_uuid(),
    nombre        varchar(100) not null,
    apellido1     varchar(100),
    apellido2     varchar(100),
    clave_issemym varchar(15)  not null,
    curp          varchar(18)  not null,

    unique (clave_issemym)
);
insert into institucion.persona
(nombre, apellido1, apellido2, clave_issemym, curp)
values
('admin', 'total', '1', '0001', 'CURP0001'),
('medico', 'base', '1', '0002', 'CURP0002');

create table institucion.rel_clinica_paciente
(
    uuid_clinica uuid not null,
    uuid_persona uuid not null,
    es_vigente bool not null,

    unique(uuid_clinica, uuid_persona),

    constraint fk_clinica foreign key (uuid_clinica)
        references institucion.clinica (uuid_clinica),
    constraint fk_persona foreign key (uuid_persona)
        references institucion.persona (uuid_persona)
);
insert into institucion.rel_clinica_paciente
(uuid_clinica, uuid_persona, es_vigente)
select
uuid_clinica, uuid_persona, true
from institucion.clinica
cross join institucion.persona
where persona.nombre = 'admin';

create table institucion.cita
(
    uuid_clinica uuid not null,
    uuid_paciente uuid not null,
    uuid_medico uuid not null,
    fecha_cita timestamptz not null,

    constraint fk_clinica foreign key (uuid_clinica)
        references institucion.clinica (uuid_clinica),
    constraint fk_paciente foreign key (uuid_paciente)
        references institucion.persona (uuid_persona),
    constraint fk_medico foreign key (uuid_medico)
        references institucion.persona (uuid_persona)
);
insert into institucion.cita
(uuid_clinica, uuid_paciente, uuid_medico, fecha_cita)
values
((select uuid_clinica from institucion.clinica limit 1),
 (select uuid_persona from institucion.persona where nombre = 'admin' limit 1),
 (select uuid_persona from institucion.persona where nombre = 'medico' limit 1),
 now() + interval '1 day');


create table sesion.usuario
(
    uuid_persona uuid primary key,
    email        varchar(200) unique not null,
    password     text                not null,

    constraint fk_persona foreign key (uuid_persona)
        references institucion.persona (uuid_persona)
);

insert into sesion.usuario
(uuid_persona, email, password)
select uuid_persona, nombre || '@gmail.com', '123456'
from institucion.persona;


create table sesion.cat_accion
(
    nombre  varchar(50) primary key
);
insert into sesion.cat_accion
(nombre)
values
('modificar');

create table sesion.recurso
(
    nombre      varchar(50) primary key,
    descripcion text
);
insert into sesion.recurso
(nombre)
values
('institucion.cita');

create table sesion.perfil
(
    nombre varchar(50) primary key
);
insert into sesion.perfil
(nombre)
values
('anonimo'),
('superadmin'),
('medico'),
('enfermeria'),
('recepcion'),
('adminclinica');

create table sesion.politica
(
    nombre      varchar(150) primary key,
    descripcion text         not null default '',
    perfil      varchar(50)  not null,
    recurso     varchar(50)  not null,
    accion      varchar(50)  not null,
    reglas      jsonb        not null,

    constraint fk_perfil foreign key (perfil)
        references sesion.perfil (nombre),
    constraint fk_recurso foreign key (recurso)
        references sesion.recurso (nombre),
    constraint fk_accion foreign key (accion)
        references sesion.cat_accion (nombre)
);
create index idx_politica_perfil on sesion.politica (perfil);
create index idx_politica_recurso on sesion.politica (recurso);
create index idx_politica_accion on sesion.politica (accion);

create table sesion.superadmin
(
    uuid_persona uuid primary key,

    constraint fk_persona foreign key (uuid_persona)
        references institucion.persona (uuid_persona)
);
insert into sesion.superadmin
(uuid_persona)
select uuid_persona from institucion.persona where nombre = 'admin' limit 1;

insert into sesion.politica
(nombre, perfil, recurso, accion, reglas)
values
('admin_cita_modificar', 'superadmin', 'institucion.cita', 'modificar',
'[{"operador": "=",
   "argumento1": {"llave": "Sesion.Perfil", "origen": "sujeto", "valor": ""},
   "argumento2": {"llave": "string", "origen": "escalar", "valor": "superadmin"}}]'),
('medico_cita_modificar', 'medico', 'institucion.cita', 'modificar',
 '[{"operador": "=",
   "argumento1": {"llave": "Sesion.UuidPersona", "origen": "sujeto", "valor": ""},
   "argumento2": {"llave": "Cita.Medico.UuidPersona", "origen": "recurso", "valor": ""}},
   {"operador": "after",
     "argumento1": {"llave": "Cita.Fecha", "origen": "recurso", "valor": ""},
     "argumento2": {"llave": "ahora", "origen": "escalar", "valor": ""}}]');


create table if not exists sesion.sesion
(
    uuid_sesion           uuid primary key                  default gen_random_uuid(),
    uuid_persona          uuid,
    tipo_sesion           varchar(20)              not null default 'cookie',
    CHECK (tipo_sesion IN ('cookie', 'jwt')),
    mapa_structs          jsonb                    not null,
    perfil                varchar(50)              not null,

    fecha_creacion        timestamp with time zone not null default now(),
    fecha_actualizacion   timestamp with time zone not null default now(),
    fecha_ultimo_acceso   timestamp with time zone not null default now(),
    fecha_expiracion      timestamp with time zone,

    direccion_ip          varchar(45),
    agente_usuario        text,
    mensajes_flash        jsonb[],

    huella                varchar(36) not null default gen_random_uuid(),
    info_dispositivo      jsonb                             default '{}'::jsonb,
    recuerdame            bool                              default false,

    -- para limitar accesos
    total_peticiones      int                               default 0,
    fecha_ultima_peticion timestamp with time zone default now(),

    -- refresh token jwt
    refresh_token_hash    bytea,

    razon_i               varchar(40),
    razon_m               varchar(40),
    fecha_i               timestamp with time zone          default now(),
    fecha_m               timestamp with time zone,

    constraint fk_perfil foreign key (perfil)
        references sesion.perfil (nombre),
    constraint fk_persona foreign key (uuid_persona)
        references institucion.persona (uuid_persona)
);

create index idx_sesion_persona on sesion.sesion (uuid_persona);
create index idx_sesion_huella on sesion.sesion (huella);
create index idx_sesion_tipo on sesion.sesion (tipo_sesion);
