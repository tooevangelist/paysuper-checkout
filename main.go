package main

import (
	"github.com/paysuper/paysuper-checkout/cmd/http"
	"github.com/paysuper/paysuper-checkout/cmd/root"
)

func main() {
	args := []string{
		"http", "-c", "configs/local.yaml", "-d",
	}
	root.ExecuteDefault(args, http.Cmd)
}
