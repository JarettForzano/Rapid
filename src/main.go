package main

import (
	"fmt"
	"os"
	"path/filepath"

	custom "github.com/Zaikoa/rapid/src/handling"
	"github.com/Zaikoa/rapid/src/transaction"

	database "github.com/Zaikoa/rapid/src/api"
	"github.com/Zaikoa/rapid/src/cloud"

	"github.com/jedib0t/go-pretty/table"
	"github.com/urfave/cli/v2"
)

// displays friends list
func displayFriends(user int) error {
	friendsList, err := database.GetFriendsList(user)
	if err != nil {
		return err
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Friend Code"})
	for _, friend := range friendsList {
		fmt.Println(friend.Name, friend.FriendCode)
		t.AppendRows([]table.Row{
			{friend.Name, friend.FriendCode},
		})
	}
	t.Render()
	return nil
}

// displays inbox
func displayInbox(user int) error {
	inbox, err := database.GetPendingTransfers(user)
	if err != nil {
		return err
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"From", "File Name"})
	for _, transaction := range inbox {
		t.AppendRows([]table.Row{
			{transaction.From_user, transaction.File_name},
		})
	}
	t.Render()
	return nil
}

func appStartup() {
	var user int
	var name string
	userDir, _ := os.UserHomeDir()

	app := &cli.App{
		Before: func(c *cli.Context) error {
			id, err := database.SetActiveSession()
			if err != nil {
				return err
			}
			if c.Args().Get(1) != "login" && id == 0 {
				return custom.NOTLOGGEDIN
			}
			if id != 0 {
				current_user, err := database.GetUserNameByID(id)
				if err != nil {
					return err
				}
				name = current_user
			}
			user = id
			return nil
		},
		After: func(c *cli.Context) error {
			if c.Bool("logout") {
				if err := database.DeactivateSession(user); err != nil {
					return err
				}
				fmt.Printf("User %s has been logged out\n", name)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "help",
				Usage: "help {Displays all commands and information}",
				Action: func(c *cli.Context) error {
					cli.ShowAppHelp(c)
					return nil
				},
			},
			{
				Name:  "user",
				Usage: "user, u {Displays user information}",
				Subcommands: []*cli.Command{
					{
						Name:    "info",
						Aliases: []string{"i"},
						Usage:   "info {Displays user information}",
						Action: func(c *cli.Context) error {
							code, err := database.GetUserFriendCode(user)
							if err != nil {
								return err
							}
							fmt.Printf("| Username   %s | Friend code   %s |\n", name, code)

							return nil
						},
					},
					{
						Name:    "login",
						Aliases: []string{"li"},
						Usage:   "login [Username] [Password] {Login to a users account}",
						Action: func(c *cli.Context) error {
							if err := database.Login(c.Args().First(), c.Args().Get(1)); err != nil {
								return err
							}
							user = database.GetCurrentId()
							current_user, err := database.GetUserNameByID(user)
							if err != nil {
								return err
							}
							name = current_user
							fmt.Printf("Currently Logged in as %s\n", name)

							return nil
						},
					},
					{
						Name:    "create",
						Aliases: []string{"c"},
						Usage:   "create [Username] [Password] {Create a users account}",
						Action: func(c *cli.Context) error {
							if err := database.CreateAccount(c.Args().First(), c.Args().Get(1)); err != nil {
								return err
							}

							if err := database.Login(c.Args().First(), c.Args().Get(1)); err != nil {
								return err
							}
							fmt.Printf("Currently Logged in as %s\n", name)

							return nil
						},
					},
					{
						Name:    "logout",
						Aliases: []string{"lo"},
						Usage:   "logout {logs current user out of their session}",
						Action: func(c *cli.Context) error {
							if err := database.DeactivateSession(user); err != nil {
								return err
							}
							fmt.Printf("User %s has been logged out\n", name)

							return nil
						},
					},
					{
						Name:    "send",
						Usage:   "send, s [User] [Filepath] {Will send user file/folder}",
						Aliases: []string{"s"},
						Action: func(c *cli.Context) error {
							if err := transaction.EncryptSend(c.Args().Get(1), user, c.Args().First()); err != nil {
								return err
							}
							fmt.Println("File has been sent and will be waiting to be accepted")

							return nil
						},
					},
				},
			},
			{
				Name:  "inbox",
				Usage: "inbox recieve, r [Filename] | inbox remove, rm [Filename] | inbox list, l {Handles inbox functionality}",
				Subcommands: []*cli.Command{
					{
						Name:    "recieve",
						Aliases: []string{"r"},
						Usage:   "inbox recieve, r [Filename] {Recieves file from inbox}",
						Action: func(c *cli.Context) error {
							if err := transaction.RecieveDecrypt(user, c.String("key"), c.Args().First(), ""); err != nil {
								return err
							}
							fmt.Println("File has been received")

							return nil
						},
					},
					{
						Name:    "remove",
						Aliases: []string{"rm"},
						Usage:   "inbox remove, rm [Filename] {Removes file from inbox}",
						Action: func(c *cli.Context) error {
							if err := cloud.DeleteFromMega(user, c.Args().First()); err != nil {
								return err
							}
							fmt.Println("File has been deleted")

							return nil
						},
					},
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "inbox list, l {Lists all messages in inbox}",
						Action: func(c *cli.Context) error {
							if err := displayInbox(user); err != nil {
								return err
							}
							fmt.Println("Inbox has been displayed")

							return nil
						},
					},
				},
			},
			{
				Name:  "friend",
				Usage: "friend add, a [Friend id] | friend remove, rm [Username] | friend list, l {Handles friend functionality}",
				Subcommands: []*cli.Command{
					{
						Name:    "add",
						Aliases: []string{"a"},
						Usage:   "friend add, a [Friend id] {Adds friend}",
						Action: func(c *cli.Context) error {
							if err := database.AddFriend(c.Args().First(), user); err != nil {
								fmt.Println(err)
							}
							fmt.Println("Friend has been added")

							return nil
						},
					},
					{
						Name:    "remove",
						Aliases: []string{"rm"},
						Usage:   "friend remove, rm [Username] {Removes friend}",
						Action: func(c *cli.Context) error {
							if err := database.DeleteFriend(user, c.Args().First()); err != nil {
								return err
							}
							fmt.Println("Friend has been deleted")

							return nil
						},
					},
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "friend list, l {Lists all friends}",
						Action: func(c *cli.Context) error {
							if err := displayFriends(user); err != nil {
								return err
							}
							fmt.Println("Friends list has been displayed")

							return nil
						},
					},
				},
			},
		},

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "key",
				Usage:       "-key, -k [Path To Key] {Specifies path to encryption key}",
				Value:       filepath.Join(userDir, "Rapid", "supersecretekey.txt"),
				Destination: &userDir,
				Aliases:     []string{"k"},
			},
			&cli.BoolFlag{
				Name:    "logout",
				Usage:   "-logout {logs user out of session after command is ran, useful if you are only running one command and do not need to be logged in for a while}",
				Aliases: []string{"lo"},
			},
		},
		EnableBashCompletion: true,
	}
	app.Suggest = true

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if err := database.InitializeDatabase(); err != nil {
		fmt.Println(err)
	}

	appStartup()

}
