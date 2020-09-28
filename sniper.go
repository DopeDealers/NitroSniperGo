package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
	gocache "github.com/patrickmn/go-cache"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	c          = gocache.New(20*time.Minute, 40*time.Minute)
	Token      string
	userID     string
	re         = regexp.MustCompile("(discord.com/gifts/|discordapp.com/gifts/|discord.gift/)([a-zA-Z0-9]+)")
	rePrivnote = regexp.MustCompile("https://privnote.com/.*")
	reGiveaway = regexp.MustCompile("You won the \\*\\*(.*)\\*\\*")
	magenta    = color.New(color.FgMagenta)
	green      = color.New(color.FgGreen)
	red        = color.New(color.FgRed)
	higreen    = color.New(color.FgHiGreen)
	strPost    = []byte("POST")
	strGet     = []byte("GET")
)

func init() {
	SetConsoleTitle("Discord Xyntix Sniper")
	file, err := ioutil.ReadFile("token.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed read file: %s\n", err)
		os.Exit(1)
	}

	var f interface{}
	err = json.Unmarshal(file, &f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse JSON: %s\n", err)
		os.Exit(1)
	}

	m := f.(map[string]interface{})

	str := fmt.Sprintf("%v", m["token"])

	flag.StringVar(&Token, "t", str, "Token")
	flag.Parse()
}

func main() {
	c := exec.Command("clear")

	c.Stdout = os.Stdout
	c.Run()
	color.Magenta(`
	▒██   ██▒▓██   ██▓ ███▄    █ ▄▄▄█████▓ ██▓▒██   ██▒     ██████  ███▄    █  ██▓ ██▓███  ▓█████  ██▀███  
	▒▒ █ █ ▒░ ▒██  ██▒ ██ ▀█   █ ▓  ██▒ ▓▒▓██▒▒▒ █ █ ▒░   ▒██    ▒  ██ ▀█   █ ▓██▒▓██░  ██▒▓█   ▀ ▓██ ▒ ██▒
	░░  █   ░  ▒██ ██░▓██  ▀█ ██▒▒ ▓██░ ▒░▒██▒░░  █   ░   ░ ▓██▄   ▓██  ▀█ ██▒▒██▒▓██░ ██▓▒▒███   ▓██ ░▄█ ▒
	 ░ █ █ ▒   ░ ▐██▓░▓██▒  ▐▌██▒░ ▓██▓ ░ ░██░ ░ █ █ ▒      ▒   ██▒▓██▒  ▐▌██▒░██░▒██▄█▓▒ ▒▒▓█  ▄ ▒██▀▀█▄  
	▒██▒ ▒██▒  ░ ██▒▓░▒██░   ▓██░  ▒██▒ ░ ░██░▒██▒ ▒██▒   ▒██████▒▒▒██░   ▓██░░██░▒██▒ ░  ░░▒████▒░██▓ ▒██▒
	▒▒ ░ ░▓ ░   ██▒▒▒ ░ ▒░   ▒ ▒   ▒ ░░   ░▓  ▒▒ ░ ░▓ ░   ▒ ▒▓▒ ▒ ░░ ▒░   ▒ ▒ ░▓  ▒▓▒░ ░  ░░░ ▒░ ░░ ▒▓ ░▒▓░
	░░   ░▒ ░ ▓██ ░▒░ ░ ░░   ░ ▒░    ░     ▒ ░░░   ░▒ ░   ░ ░▒  ░ ░░ ░░   ░ ▒░ ▒ ░░▒ ░      ░ ░  ░  ░▒ ░ ▒░
	 ░    ░   ▒ ▒ ░░     ░   ░ ░   ░       ▒ ░ ░    ░     ░  ░  ░     ░   ░ ░  ▒ ░░░          ░     ░░   ░ 
	 ░    ░   ░ ░              ░           ░   ░    ░           ░           ░  ░              ░  ░   ░     
			  ░ ░                                                                                          `)
	dg, err := discordgo.New(Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(disconnect)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	t := time.Now()
	color.HiGreen("Sniping Discord Nitro on " + strconv.Itoa(len(dg.State.Guilds)) + " Servers ₿\n\n")

	magenta.Print(t.Format("15:04:05 "))
	higreen.Print("[+]")
	fmt.Println(" Bot is ready")
	userID = dg.State.User.ID

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func checkCode(bodyString string) {
	_, _ = magenta.Print(time.Now().Format("15:04:05 "))
	if strings.Contains(bodyString, "This gift has been redeemed already.") {
		color.Yellow("[-] Code has been already redeemed")
	} else if strings.Contains(bodyString, "nitro") {
		_, _ = higreen.Println("[+] Code applied")
	} else if strings.Contains(bodyString, "Unknown Gift Code") {
		_, _ = red.Println("[x] Invalid Code")
	} else {
		color.Yellow("[-] Cannot check gift validity")
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if re.Match([]byte(m.Content)) {

		code := re.FindStringSubmatch(m.Content)
		if m.Author.ID == userID {
			magenta.Print(time.Now().Format("15:04:05 "))
			red.Print("[x] Auto-detected user sent code: ")
			color.HiMagenta(" Will not Claim/Check any further")
			return
		}
		if cached, found := c.Get("codes"); found {
			if code[2] == cached {
				magenta.Print(time.Now().Format("15:04:05 "))
				red.Print("[x] Cached/Claimed code found")
				red.Print(code)
				return
			}
		}

		if len(code) < 2 {
			return
		}

		if len(code[2]) < 16 {
			_, _ = magenta.Print(time.Now().Format("15:04:05 "))
			_, _ = red.Print("[x] Auto-detected a fake code: ")
			_, _ = red.Print(code[2])
			fmt.Println(" from " + m.Author.String())
			return
		}

		var strRequestURI = []byte("https://discordapp.com/api/v6/entitlements/gift-codes/" + code[2] + "/redeem")
		req := fasthttp.AcquireRequest()
		req.Header.SetContentType("application/json")
		req.Header.Set("authorization", Token)
		req.SetBody([]byte(`{"channel_id":` + m.ChannelID + "}"))
		req.Header.SetMethodBytes(strPost)
		req.SetRequestURIBytes(strRequestURI)
		res := fasthttp.AcquireResponse()

		if err := fasthttp.Do(req, res); err != nil {
			panic("handle error")
		}

		fasthttp.ReleaseRequest(req)

		body := res.Body()

		bodyString := string(body)
		fasthttp.ReleaseResponse(res)

		_, _ = magenta.Print(time.Now().Format("15:04:05 "))
		_, _ = green.Print("[-] Sniped code: ")
		_, _ = red.Print(code[2])
		guild, err := s.State.Guild(m.GuildID)
		if err != nil || guild == nil {
			guild, err = s.Guild(m.GuildID)
			if err != nil {
				println()
				checkCode(bodyString)
				c.Set("codes", code[2], gocache.NoExpiration)
				return
			}
		}

		channel, err := s.State.Channel(m.ChannelID)
		if err != nil || guild == nil {
			channel, err = s.Channel(m.ChannelID)
			if err != nil {
				println()
				checkCode(bodyString)
				c.Set("codes", code[2], gocache.NoExpiration)
				return
			}
		}

		print(" from " + m.Author.String())
		_, _ = magenta.Println(" [" + guild.Name + " > " + channel.Name + "]")
		checkCode(bodyString)

	}

}

func disconnect(s *discordgo.Session, d *discordgo.Disconnect) {
	// add cache saving to file soon
}

func SetConsoleTitle(title string) (int, error) {
	handle, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		return 0, err
	}
	defer syscall.FreeLibrary(handle)
	proc, err := syscall.GetProcAddress(handle, "SetConsoleTitleW")
	if err != nil {
		return 0, err
	}
	r, _, err := syscall.Syscall(proc, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), 0, 0)
	return int(r), err
}
