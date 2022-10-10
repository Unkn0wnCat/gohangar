package main

import (
	"github.com/Unkn0wnCat/gohangar/hangar"
	"github.com/urfave/cli/v2"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	app := &cli.App{
		Name:        "gohangar",
		Usage:       "HTTP directory listing",
		ArgsUsage:   "[base directory]",
		UsageText:   "gohangar [options] [command]",
		Description: "GoHangar allows you to easily expose a local directory as a HTTP download base.",
		Authors: []*cli.Author{
			{
				Name:  "Kevin Kandlbinder",
				Email: "kevin@kevink.dev",
			},
		},
		Copyright: "Copyright (C) 2022  GoHangar Authors\nThis program is free software: you can redistribute it and/or modify\nit under the terms of the GNU General Public License as published by\nthe Free Software Foundation, either version 3 of the License, or\n(at your option) any later version.\n\nThis program is distributed in the hope that it will be useful,\nbut WITHOUT ANY WARRANTY; without even the implied warranty of\nMERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the\nGNU General Public License for more details.\n\nYou should have received a copy of the GNU General Public License\nalong with this program.  If not, see <https://www.gnu.org/licenses/>.\n",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Value:   "./data",
				Usage:   "Directory to serve",
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Value:   "DLBase",
				Usage:   "Name displayed on the site",
			},
			&cli.StringFlag{
				Name:    "attribution",
				Aliases: []string{"a"},
				Usage:   "Overrides the attribution in the footer (HTML allowed)",
			},
			&cli.BoolFlag{
				Name:  "no-header",
				Value: false,
				Usage: "Disables the header",
			},
			&cli.StringFlag{
				Name:    "header",
				Aliases: []string{"i"},
				Value:   "/static/banner.jpg",
				Usage:   "Sets the path to the header image from webroot",
			},
			&cli.StringFlag{
				Name:  "app",
				Value: "GoHangar",
				Usage: "App name displayed in the footer",
			},
		},
		Action: func(cCtx *cli.Context) error {
			base := os.DirFS(cCtx.String("dir"))

			myHangar, err := hangar.New(base)
			if err != nil {
				return err
			}

			myHangar.AppName = cCtx.String("app")
			myHangar.Attribution = template.HTML(cCtx.String("attribution"))
			myHangar.NoHeader = cCtx.Bool("no-header")
			myHangar.Banner = cCtx.String("header")
			myHangar.SiteName = cCtx.String("name")

			err = http.ListenAndServe("localhost:8123", myHangar)
			if err != nil {
				return err
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
