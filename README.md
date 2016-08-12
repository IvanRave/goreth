# Golang Redis authentication methods

Generates and validates a verification code for a user login: email or phone number.


## Usage

```
import "github.com/ivanrave/goreth"

vcode := "any generated verification code, like random numbers"

// save a verification code with 60 second expiration
err := goreth.SetLoginAndVcode("myemail", vcode, 60)

// retrieve a verification code for required email
vcode, retry, err := goreth.GetVcode("myemail")

// if a verification code is wrong - add a new retry
err := goreth.AddRetry("myemail")

// if count of retries more than max count - delete a key
err := goreth.DelKey("myemail")
``


## Steps

1. Enter a email or phone number
2. Receive a verification code by email or phone, eg. 5 digits
3. Enter a email/phone + code
4. Authentication token is generated and saved to a client's browser
5. A user can view secured pages


## Stack

go-redis: https://github.com/go-redis/redis