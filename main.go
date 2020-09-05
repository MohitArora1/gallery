package main

import (
	"github.com/MohitArora1/gallery/controller"
	"github.com/MohitArora1/gallery/utils"
)

func main() {
	utils.LoadConfig()
	controller.RunController(":8080")
}
