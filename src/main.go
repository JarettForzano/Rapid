package main

import (
	"fmt"
	"os"
	"path/filepath"

	custom "github.com/Zaikoa/rapid/src/handling"

	database "github.com/Zaikoa/rapid/src/api"
	"github.com/Zaikoa/rapid/src/cloud"

	"github.com/jedib0t/go-pretty/table"
	"github.com/urfave/cli/v2"
)

// displays friends list
func displayFriends(user int) error {
	if user == 0 {
		return custom.NewError("User must be logged in to use this method")
	}

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
	if user == 0 {
		return custom.NewError("User must be logged in to use this method")
	}

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
	userDir, _ := os.UserHomeDir()

	app := &cli.App{
		Before: func(c *cli.Context) error {
			err := database.SetActiveSession()
			if err != nil {
				return err
			}
			user = database.GetCurrentId()
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "login",
				Usage: "login [Username] [Password] {Login to a users account}",
				Action: func(c *cli.Context) error {
					err := database.Login(c.Args().First(), c.Args().Get(1))
					if err != nil {
						return err
					}
					user = database.GetCurrentId()
					fmt.Printf("Currently Logged in as %s\n", database.GetUserNameByID(user))

					return nil
				},
			},
			{
				Name:  "logout",
				Usage: "logout {logs current user out of their session}",
				Action: func(c *cli.Context) error {
					err := database.DeactivateSession(user)
					if err != nil {
						return err
					}
					fmt.Printf("User %s has been logged out\n", database.GetUserNameByID(user))

					return nil
				},
			},
			{
				Name:  "create",
				Usage: "create [Username] [Password] {Create a users account}",
				Action: func(c *cli.Context) error {
					err := database.CreateAccount(c.Args().First(), c.Args().Get(1))
					if err != nil {
						return err
					}

					err = database.Login(c.Args().First(), c.Args().Get(1))
					if err != nil {
						return err
					}

					user = database.GetCurrentId()
					fmt.Printf("Currently Logged in as %s\n", database.GetUserNameByID(user))

					return nil
				},
			},
			{
				Name:  "help",
				Usage: "help {Displays all commands and information}",
				Action: func(c *cli.Context) error {
					cli.ShowAppHelp(c)
					return nil
				},
			},
			{
				Name:    "user",
				Usage:   "user, u {Displays user information}",
				Aliases: []string{"u"},
				Action: func(c *cli.Context) error {
					result, err := database.GetUserFriendCode(user)
					if err != nil {
						return err
					}
					fmt.Printf("| Username   %s | Friend code   %s |\n", database.GetUserNameByID(user), result)
					return nil
				},
			},
			{
				Name:    "send",
				Usage:   "send, s [User] [Filepath] {Will send user file/folder}",
				Aliases: []string{"s"},
				Action: func(c *cli.Context) error {
					err := cloud.UploadToMega(c.Args().Get(1), user, c.Args().First())
					if err != nil {
						return err
					}
					fmt.Println("File has been sent and will be waiting to be accepted")
					return nil
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
							fmt.Println("Key is:", c.String("key"))
							err := cloud.DownloadFromMega(user, c.Args().First(), "", c.String("key"))
							if err != nil {
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
							result, err := cloud.DeleteFromMega(user, c.Args().First())
							if err != nil {
								fmt.Println(err)
							}
							if result {
								fmt.Println("File has been deleted")
							} else {
								fmt.Println("Could not delete the file from the inbox")
							}
							return nil
						},
					},
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "inbox list, l {Lists all messages in inbox}",
						Action: func(c *cli.Context) error {
							err := displayInbox(user)
							if err != nil {
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
							result, err := database.AddFriend(c.Args().First(), user)
							if err != nil {
								fmt.Println(err)
							}

							if !result {
								fmt.Println("There does not exist a user with that friend code.")

							}

							if result {
								fmt.Println("Friend has been added")
							}
							return nil
						},
					},
					{
						Name:    "remove",
						Aliases: []string{"rm"},
						Usage:   "friend remove, rm [Username] {Removes friend}",
						Action: func(c *cli.Context) error {
							result, err := database.DeleteFriend(user, c.Args().First())
							if err != nil {
								fmt.Println("Failed to remove friend", err)
							}
							if result {
								fmt.Println("Friend has been deleted")
							}
							return nil
						},
					},
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "friend list, l {Lists all friends}",
						Action: func(c *cli.Context) error {
							err := displayFriends(user)
							if err != nil {
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

// Main method for runnning the system
func main() {
	err := database.InitializeDatabase()
	if err != nil {
		fmt.Println(err)
	}
	appStartup()
}
