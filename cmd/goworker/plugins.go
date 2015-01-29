package main

import (
	_ "github.com/eliothedeman/workhorse/provider/http"
	_ "github.com/eliothedeman/workhorse/provider/redis"
	_ "github.com/eliothedeman/workhorse/worker/cli"
	_ "github.com/eliothedeman/workhorse/worker/disk"
	_ "github.com/eliothedeman/workhorse/worker/http"
)
