package cl

import (
	"fmt"
	"l0/internal/store"
	"os"
)

type CL struct {
	DB *store.Store
}

var command_list string = `Command list:
list: show all orders in Store
help: get help
exit: exit
`

func (c *CL) Start() {
	fmt.Println("Service is working")
	fmt.Println(command_list)
	c.loopWorker()
}

func (c *CL) loopWorker() {
	var cmd string

	for {
		fmt.Print("> ")
		fmt.Scan(&cmd)

		switch cmd {
		case "list":
			fmt.Println("Orders:")
			for _, ord := range *cl.DB.Cache {
				fmt.Println("---", ord.OrderUID)
			}
		case "help":
			fmt.Println(command_list)
		case "exit":
			fmt.Println("Program closing")
			os.Exit(0)
		default:
			fmt.Println("Command not found")
			fmt.Println(command_list)
		}
	}
}

func New(db *store.Store) *CL {
	return &CL{
		DB: db,
	}
}
