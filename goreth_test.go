package goreth

import (
	"log"
	"fmt"
	"time"
)

// Integration tests using expired values only
// Delete a key after test
func ExampleDbauth(){
	// todo: use env variables for db config
	err := InitPool("127.0.0.1", "6379", 0, "")

	if err != nil {
		log.Fatal(err)
	}

	key := "qwe"
	var scnd time.Duration = 1

	err = SetLoginAndVcode(key, "121", scnd)

	if err != nil {	log.Fatal(err)	}

	err = SetLoginAndVcode(key, "321", scnd)

	if err != nil {
		if err == ErrLgnExists {
			fmt.Println("exists")
		} else {
			log.Fatal(err)
		}
	}

	// 50k - 6.2s HIncrBy (clean)
	// 50k - 7.1s HIncrBy (eval)
	// 50k - 6.7s HIncrBy (eval + EXISTS)
	for i := 0; i < 8; i++ {
		err = AddRetry(key)
		if err != nil {	log.Fatal(err)	}
	}

	vcode, retry, err := GetVcode(key)

	if err != nil {	log.Fatal(err)	}
	
	fmt.Println(vcode, retry)

	err = DelKey(key)

	if err != nil {	log.Fatal(err)	}

	// Output:
	// exists
	// 121 8
}
