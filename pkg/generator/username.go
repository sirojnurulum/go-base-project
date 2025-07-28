package generator

import (
	"fmt"
	"strings"
	"time"
)

func GenerateFromEmail(email string) string {
	return fmt.Sprintf("%s_%d", strings.Split(email, "@")[0], time.Now().Unix())
}
