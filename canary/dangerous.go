package canary

import (
	"fmt"
	"os"
)

func AcknowledgeDangerousCanary() error {
	if _, ok := os.LookupEnv("ZROK_DANGEROUS_CANARY"); !ok {
		return fmt.Errorf("this is a dangerous canary; see canary docs for details on enabling")
	}
	return nil
}
