document.addEventListener('DOMContentLoaded', () => {
	document.addEventListener('click', (event) => {
		const button = event.target.closest('.js-export-action');
		if (!button) {
			return;
		}

		const exportUrl = button.dataset.exportUrl;
		const toastUrl = button.dataset.toastUrl;

		if (!exportUrl || !toastUrl) {
			return;
		}

		window.open(exportUrl);

		if (window.htmx) {
			setTimeout(() => {
				window.htmx.ajax('GET', toastUrl, { target: '#toast-container', swap: 'beforeend' });
			}, 100);
		}
	});
});
