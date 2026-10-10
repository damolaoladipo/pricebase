package utils

import (
	"encoding/json"
	"fmt"
)

func ConsoleLog(messages ...any) {
	for _, message := range messages {
		switch m := message.(type) {

		case string:
			fmt.Print(m)

		case error:
			fmt.Print(m.Error())

		case []byte:
			fmt.Print(string(m))

		default:
			data, err := json.MarshalIndent(m, "", "  ")

			if err != nil {
				fmt.Printf("%+v", m)
				continue
			}

			fmt.Print(string(data))
		}

		fmt.Print(" ")
	}

	fmt.Println()
}
