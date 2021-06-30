package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/ahmetb/go-linq/v3"
	"github.com/blinkbean/dingtalk"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/urfave/cli/v2"
	"gopkg.in/fatih/set.v0"
)

const (
	dingToken    = "--your dingToken--"
	secretString = "--your secret--"
)

var (
	yamlUri = "./db.yaml"
	//静默时间
	forTime = time.Duration(time.Minute * 1)

	errList = []string{
		"retrying",
		"abort",
	}

	preString = "l2"
)

func main() {
	StartCmd()
	// StartAction(nil)
}

func ReportFunc(sConfig set.Interface) {

	var originConfig = &DataConfig{}
	var finnalConfig = &DataConfig{}

	originConfig = readYaml(yamlUri)

	originsConfig := set.New(set.ThreadSafe)

	for _, item := range originConfig.Data {
		originsConfig.Add(item)
	}

	newsConfig := sConfig.Copy()

	newConfig := DataConfig{}

	sizet := newsConfig.Size()
	for i := 0; i <= sizet; i++ {
		if newsConfig.Size() == 0 {
			break
		}
		newConfig.Data = append(newConfig.Data, newsConfig.Pop().(Config))
	}

	cli := dingtalk.InitDingTalkWithSecret(dingToken, secretString)

	sentMsg := ""

	originNames := []string{}

	From(originConfig.Data).SelectT(func(item Config) string {
		return item.Name
	}).ToSlice(&originNames)

	finnalSet := set.New(set.ThreadSafe)
	sentSet := set.New(set.ThreadSafe)

	//发出超过静默时间的通知
	for _, item := range newConfig.Data {

		if contains(originNames, item.Name) {
			oc := From(originConfig.Data).WhereT(func(c Config) bool {
				return c.Name == item.Name
			}).First()

			oct := oc.(Config)
			if forTime > time.Now().Sub(oct.EventTime) {
				finnalSet.Add(oct)
			} else {
				oct.EventTime = time.Now()
				sentSet.Add(item)
				finnalSet.Add(item)
			}
		} else {
			sentSet.Add(item)
			finnalSet.Add(item)
		}
	}

	tmpMsg := ""

	ssize := sentSet.Size()
	for i := 0; i <= ssize; i++ {
		item := sentSet.Pop()
		if item == nil {
			break
		}
		tmpMsg += fmt.Sprintf("%s error \n", item.(Config).Name)
	}

	if tmpMsg != "" {
		msg := fmt.Sprintf("------start %s %s------\n%s------end------\n", preString, time.Now().Format("2006-01-02 15:04:05"), tmpMsg)
		sentMsg = msg
		fmt.Println(sentMsg)

	}

	fsize := finnalSet.Size()
	for i := 0; i <= fsize; i++ {
		item := finnalSet.Pop()
		if item == nil {
			break
		}
		finnalConfig.Data = append(finnalConfig.Data, item.(Config))
	}

	fmt.Printf("record yaml：%v", finnalConfig)
	writeYaml(yamlUri, *finnalConfig)

	e := cli.SendTextMessage(sentMsg)
	if e != nil {
		fmt.Println(e.Error())
	}
}

func StartCmd() {
	app := &cli.App{
		Name:  "dm",
		Usage: "docker monitor",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "f",
				Usage: "how many mins for silence",
				Value: "1",
			},
			&cli.StringFlag{
				Name:  "k",
				Usage: "keyworks for identify split by ','",
				Value: "retrying,abort",
			},
			&cli.StringFlag{
				Name:  "p",
				Usage: "prefix",
				Value: "l2",
			},
			&cli.StringFlag{
				Name:  "y",
				Usage: "locate db.yaml",
				Value: "./db.yaml",
			},
		},
		Action: StartAction,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func StartAction(c *cli.Context) error {
	// fstr := "1440"
	// kstr := "ffffffff"
	// pstr := "l2"
	// ystr := "F:\\Go\\src\\dockermonitor\\db.yaml"

	fstr := c.String("f")
	kstr := c.String("k")
	pstr := c.String("p")
	ystr := c.String("y")

	yamlUri = ystr
	preString = pstr
	kwords := strings.Split(kstr, ",")

	errList = kwords

	f, err := strconv.ParseInt(fstr, 10, 64)

	if err != nil {
		return err
	}

	forTime = time.Duration(int64(time.Minute) * f)

	sConfig := set.New(set.ThreadSafe)

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {

		if container.State == "exited" {
			t := time.Now()

			sConfig.Add(Config{Name: container.Names[0], EventTime: t})

			fmt.Printf("rev error：name %v time %v \n", container.Names[0], t.Format("2006-01-02 15:04:05"))

			continue
		}

		io, err := cli.ContainerLogs(ctx, container.ID, types.ContainerLogsOptions{Tail: "1", ShowStdout: true})

		if err != nil {
			fmt.Println(err.Error())
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(io)
		newStr := strings.ToLower(buf.String())

		for _, v := range errList {
			if strings.Contains(newStr, v) {

				t := time.Now()
				sConfig.Add(Config{Name: container.Names[0], EventTime: t})

				fmt.Printf("rev error：name %v time %v \n", container.Names[0], t.Format("2006-01-02 15:04:05"))

				goto jump
			}
		}

	jump:
		continue
	}

	ReportFunc(sConfig)

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
