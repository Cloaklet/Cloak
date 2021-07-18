package server

import "embed"

//go:embed dist/*
var frontend embed.FS
