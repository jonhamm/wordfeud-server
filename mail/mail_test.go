package mail

import (
	"testing"

	"github.com/joho/godotenv"
)

func Test_SendMail(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Errorf("Test_SendMail - godotenv.Load() failed \n%s", err.Error())
	}

	err = SendMail()
	if err != nil {
		t.Errorf("Test_SendMail - SendMail failed \n%s", err.Error())
	}
}
