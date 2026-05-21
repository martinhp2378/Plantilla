package sesiones

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.Getenv("PATH"))
}
