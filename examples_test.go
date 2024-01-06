package flago

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

type MyFlags struct {
	Verbose   int      `flags:"verbose,v" flagUsage:"verbose mode: 1,2,3"`
	Login     *string  `flag:"login" flagUsage:"user login"`
	FileNames []string `flagArgs:"true"`
}

func TestMyFlagsExample1(t *testing.T) {
	flagSet := NewFlagSet("myApp", flag.ExitOnError)
	myFlags := MyFlags{}
	if err := flagSet.StructVar(&myFlags); err != nil {
		panic(err)
	}
	if err := flagSet.Parse([]string{"-v", "2", "--login", "user1", "file1", "file2"}); err != nil {
		panic(err)
	}
	require.Equal(t, 2, myFlags.Verbose)
	require.Equal(t, "user1", *myFlags.Login)
	require.ElementsMatch(t, []string{"file1", "file2"}, myFlags.FileNames)
}

func TestMyFlagsExample2(t *testing.T) {
	flagSet := NewFlagSet("myApp", flag.ExitOnError)
	myFlags := MyFlags{}
	if err := flagSet.StructVar(&myFlags); err != nil {
		panic(err)
	}
	if err := flagSet.Parse([]string{"--verbose", "2"}); err != nil {
		panic(err)
	}
	require.Equal(t, 2, myFlags.Verbose)
	require.Nil(t, myFlags.Login)
	require.Empty(t, myFlags.FileNames)
}

func TestMyFlagsExample3(t *testing.T) {
	flagSet := NewFlagSet("myApp", flag.ExitOnError)
	defaultLogin := "admin"
	myFlags := MyFlags{
		Verbose: 1,
		Login:   &defaultLogin,
	}
	if err := flagSet.StructVar(&myFlags); err != nil {
		panic(err)
	}
	if err := flagSet.Parse([]string{"file1", "file2"}); err != nil {
		panic(err)
	}
	require.Equal(t, 1, myFlags.Verbose)
	require.Equal(t, "admin", *myFlags.Login)
	require.ElementsMatch(t, []string{"file1", "file2"}, myFlags.FileNames)
}

func TestMyFlagsExample4(t *testing.T) {
	type PersonFlags struct {
		Name  string `flag:"name" flagUsage:"person name"`
		Email string `flag:"email" flagUsage:"person email"`
	}
	type MyParentFlags struct {
		Sender   PersonFlags `flagPrefix:"sender-" flagUsagePrefix:"sender "`
		Receiver PersonFlags `flagPrefix:"receiver-" flagUsagePrefix:"receiver "`
	}

	flagSet := NewFlagSet("myApp", flag.ExitOnError)
	myFlags := MyParentFlags{}
	if err := flagSet.StructVar(&myFlags); err != nil {
		panic(err)
	}
	if err := flagSet.Parse([]string{
		"--sender-name", "John",
		"--sender-email", "j@email.com",
		"--receiver-name", "Dave",
		"--receiver-email", "d@email.com",
	}); err != nil {
		panic(err)
	}
	require.Equal(t, "John", myFlags.Sender.Name)
	require.Equal(t, "j@email.com", myFlags.Sender.Email)
	require.Equal(t, "Dave", myFlags.Receiver.Name)
	require.Equal(t, "d@email.com", myFlags.Receiver.Email)
}
