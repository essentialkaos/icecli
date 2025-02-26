package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/fmtutil"
	"github.com/essentialkaos/ek/v13/fmtutil/table"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/support"
	"github.com/essentialkaos/ek/v13/support/deps"
	"github.com/essentialkaos/ek/v13/support/pkgs"
	"github.com/essentialkaos/ek/v13/terminal"
	"github.com/essentialkaos/ek/v13/terminal/tty"
	"github.com/essentialkaos/ek/v13/timeutil"
	"github.com/essentialkaos/ek/v13/usage"
	"github.com/essentialkaos/ek/v13/usage/completion/bash"
	"github.com/essentialkaos/ek/v13/usage/completion/fish"
	"github.com/essentialkaos/ek/v13/usage/completion/zsh"
	"github.com/essentialkaos/ek/v13/usage/man"
	"github.com/essentialkaos/ek/v13/usage/update"

	ic "github.com/essentialkaos/go-icecast/v3"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "icecli"
	DESC = "Icecast CLI"
	VER  = "1.1.1"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	CMD_HELP         = "help"
	CMD_STATS        = "stats"
	CMD_KILL_CLIENT  = "kill-client"
	CMD_KILL_SOURCE  = "kill-source"
	CMD_LIST_CLIENTS = "list-clients"
	CMD_LIST_MOUNTS  = "list-mounts"
	CMD_MOVE_CLIENTS = "move-clients"
	CMD_UPDATE_META  = "update-meta"
)

const (
	OPT_HOST     = "H:host"
	OPT_USER     = "U:user"
	OPT_PASS     = "P:password"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"

	OPT_VERB_VER     = "vv:verbose-version"
	OPT_COMPLETION   = "completion"
	OPT_GENERATE_MAN = "generate-man"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// optMap is map with options
var optMap = options.Map{
	OPT_HOST:     {Value: "http://127.0.0.1:8000", Alias: "url"},
	OPT_USER:     {Value: "admin", Alias: "login"},
	OPT_PASS:     {Value: "hackme", Alias: "pass"},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL},
	OPT_VER:      {Type: options.MIXED},

	OPT_VERB_VER:     {Type: options.BOOL},
	OPT_COMPLETION:   {},
	OPT_GENERATE_MAN: {Type: options.BOOL},
}

// colorTagApp contains color tag for app name
var colorTagApp string

// colorTagVer contains color tag for app version
var colorTagVer string

// client is icecast API client
var client *ic.API

// ////////////////////////////////////////////////////////////////////////////////// //

// Run is main application function
func Run(gitRev string, gomod []byte) {
	preConfigureUI()

	args, errs := options.Parse(optMap)

	if !errs.IsEmpty() {
		terminal.Error("Options parsing errors:")
		terminal.Error(errs.Error(" - "))
		os.Exit(1)
	}

	configureUI()

	switch {
	case options.Has(OPT_COMPLETION):
		os.Exit(printCompletion())
	case options.Has(OPT_GENERATE_MAN):
		printMan()
		os.Exit(0)
	case options.GetB(OPT_VER):
		genAbout(gitRev).Print(options.GetS(OPT_VER))
		os.Exit(0)
	case options.GetB(OPT_VERB_VER):
		support.Collect(APP, VER).
			WithRevision(gitRev).
			WithDeps(deps.Extract(gomod)).
			WithPackages(pkgs.Collect("icecast,icecast2,icecast-kh")).
			Print()
		os.Exit(0)
	case len(args) == 0, options.GetB(OPT_HELP):
		genUsage().Print()
		os.Exit(0)
	}

	if args.Get(0).ToLower().String() == CMD_HELP {
		checkForRequiredArgs(args, 1)
		showHelp(args.Get(0).String())
	} else {
		execCommand(args)
	}
}

// preConfigureUI preconfigures UI based on information about user terminal
func preConfigureUI() {
	if !tty.IsTTY() {
		fmtc.DisableColors = true
	}

	fmtutil.SeparatorSymbol = "–"
	fmtutil.SeparatorFullscreen = true
	fmtutil.SizeSeparator = " "
	table.SeparatorSymbol = "–"
	table.HeaderCapitalize = true

	switch {
	case fmtc.Is256ColorsSupported():
		colorTagApp, colorTagVer = "{*}{#27}", "{#27}"
	default:
		colorTagApp, colorTagVer = "{*}{b}", "{b}"
	}
}

// configureUI configures user interface
func configureUI() {
	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}
}

// execCommand executes command
func execCommand(args options.Arguments) {
	var err error

	client, err = ic.NewAPI(
		options.GetS(OPT_HOST),
		options.GetS(OPT_USER),
		options.GetS(OPT_PASS),
	)

	if err != nil {
		printErrorExit(err.Error())
	}

	cmd := args.Get(0).ToLower().String()

	switch cmd {
	case CMD_STATS:
		showServerStats()
	case CMD_LIST_MOUNTS:
		listMounts()
	case CMD_LIST_CLIENTS:
		checkForRequiredArgs(args, 1)
		listClients(args.Get(1).String())
	case CMD_MOVE_CLIENTS:
		checkForRequiredArgs(args, 2)
		moveClients(
			args.Get(1).String(),
			args.Get(2).String(),
		)
	case CMD_UPDATE_META:
		checkForRequiredArgs(args, 3)
		updateMeta(
			args.Get(1).String(),
			args.Get(2).String(),
			args.Get(3).String(),
		)
	case CMD_KILL_CLIENT:
		checkForRequiredArgs(args, 2)
		killClient(
			args.Get(1).String(),
			args.Get(2).String(),
		)
	case CMD_KILL_SOURCE:
		checkForRequiredArgs(args, 1)
		killSource(args.Get(1).String())
	default:
		terminal.Error("Unknown or unsupported command %q", cmd)
		os.Exit(1)
	}
}

// showHelp prints command usage info
func showHelp(command string) {
	switch command {
	case CMD_STATS:
		helpCmdStats()
	case CMD_LIST_MOUNTS:
		helpCmdListMounts()
	case CMD_LIST_CLIENTS:
		helpCmdListClients()
	case CMD_MOVE_CLIENTS:
		helpCmdMoveClients()
	case CMD_UPDATE_META:
		helpCmdUpdateMeta()
	case CMD_KILL_CLIENT:
		helpCmdKillClient()
	case CMD_KILL_SOURCE:
		helpCmdKillSource()
	default:
		genUsage().Print()
	}
}

// showServerStats prints server stats
func showServerStats() {
	stats, err := client.GetStats()

	if err != nil {
		printErrorExit(err.Error())
	}

	fmtc.NewLine()
	printServerHeader(stats.Info.ID)
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Sources", fmtutil.PrettyNum(stats.Stats.Sources))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Banned IPs", fmtutil.PrettyNum(stats.Stats.BannedIPs))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Clients", fmtutil.PrettyNum(stats.Stats.Clients))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Connections", fmtutil.PrettyNum(stats.Stats.Connections))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Listeners", fmtutil.PrettyNum(stats.Stats.Listeners))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Stats", fmtutil.PrettyNum(stats.Stats.Stats))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Client Connections", fmtutil.PrettyNum(stats.Stats.ClientConnections))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "File Connections", fmtutil.PrettyNum(stats.Stats.FileConnections))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Listener Connections", fmtutil.PrettyNum(stats.Stats.ListenerConnections))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Stats Connections", fmtutil.PrettyNum(stats.Stats.StatsConnections))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Source Client Connections", fmtutil.PrettyNum(stats.Stats.SourceClientConnections))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Source Relay Connections", fmtutil.PrettyNum(stats.Stats.SourceRelayConnections))
	fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Source Total Connections", fmtutil.PrettyNum(stats.Stats.SourceTotalConnections))
	fmtc.Printfn(
		" {*}%-28s{!} {s}|{!} %s {s-}(%s){!}", "Stream Bytes Read",
		fmtutil.PrettyNum(stats.Stats.StreamBytesRead),
		fmtutil.PrettySize(stats.Stats.StreamBytesRead),
	)
	fmtc.Printfn(
		" {*}%-28s{!} {s}|{!} %s {s-}(%s){!}", "Stream Bytes Sent",
		fmtutil.PrettyNum(stats.Stats.StreamBytesSent),
		fmtutil.PrettySize(stats.Stats.StreamBytesSent),
	)

	for path, source := range stats.Sources {
		showSeparator(false)
		fmtc.Printfn(" {*y}%s{!} {s-}(online: %s){!}", path, timeutil.PrettyDuration(time.Since(source.StreamStarted)))
		showSeparator(false)
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Source IP", source.SourceIP)
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Name", formatString(source.Info.Name))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Genre", formatString(source.Genre))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Description", formatString(source.Info.Description))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Type", formatString(source.Info.Type))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "URL", formatString(source.Info.URL))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Listen URL", formatString(source.ListenURL))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "SubType", formatString(source.Info.SubType))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %t", "Public", source.Public)
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "User-Agent", formatString(source.UserAgent))
		showSeparator(true)
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Bitrate", fmtutil.PrettyNum(source.AudioInfo.Bitrate))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Channels", fmtutil.PrettyNum(source.AudioInfo.Channels))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s Hz", "SampleRate", fmtutil.PrettyNum(source.AudioInfo.SampleRate))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "CodecID", fmtutil.PrettyNum(source.AudioInfo.CodecID))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "RawInfo", formatString(source.AudioInfo.RawInfo))
		showSeparator(true)
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Artist", formatString(source.Track.Artist))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Title", formatString(source.Track.Title))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Artwork", formatString(source.Track.Artwork))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Metadata URL", formatString(source.Track.MetadataURL))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "RawInfo", formatString(source.Track.RawInfo))
		fmtc.Printfn(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s ago){!}", "Metadata Updated",
			timeutil.Format(source.MetadataUpdated, "%Y/%m/%d %H:%M:%S"),
			timeutil.PrettyDuration(time.Since(source.MetadataUpdated)),
		)
		showSeparator(true)
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Listeners", fmtutil.PrettyNum(source.Stats.Listeners))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Listener Peak", fmtutil.PrettyNum(source.Stats.ListenerPeak))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Max Listeners", fmtutil.PrettyNum(source.Stats.MaxListeners))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Slow Listeners", fmtutil.PrettyNum(source.Stats.SlowListeners))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Listener Connections", fmtutil.PrettyNum(source.Stats.ListenerConnections))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Connected", fmtutil.PrettyNum(source.Stats.Connected))
		fmtc.Printfn(" {*}%-28s{!} {s}|{!} %s", "Queue Size", fmtutil.PrettyNum(source.Stats.QueueSize))

		fmtc.Printfn(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s/s){!}", "Incoming Bitrate",
			fmtutil.PrettyNum(source.Stats.IncomingBitrate),
			fmtutil.PrettySize(source.Stats.IncomingBitrate),
		)

		fmtc.Printfn(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s/s){!}", "Outgoing Bitrate",
			fmtutil.PrettyNum(source.Stats.OutgoingBitrate),
			fmtutil.PrettySize(source.Stats.OutgoingBitrate),
		)

		fmtc.Printfn(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s){!}", "Total Bytes Read",
			fmtutil.PrettyNum(source.Stats.TotalBytesRead),
			fmtutil.PrettySize(source.Stats.TotalBytesRead),
		)

		fmtc.Printfn(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s){!}", "Total Bytes Sent",
			fmtutil.PrettyNum(source.Stats.TotalBytesSent),
			fmtutil.PrettySize(source.Stats.TotalBytesSent),
		)
	}

	showSeparator(false)
	fmtc.NewLine()
}

// listMounts prints list of all connected mount points
func listMounts() {
	mounts, err := client.ListMounts()

	if err != nil {
		printErrorExit(err.Error())
	}

	if len(mounts) == 0 {
		fmtc.Println("{y}No mounts found{!}")
		return
	}

	t := table.NewTable("path", "listeners", "connected", "content-type")
	t.SetAlignments(table.ALIGN_LEFT, table.ALIGN_RIGHT, table.ALIGN_RIGHT)
	t.SetSizes(20, 10, 10)

	fmtc.NewLine()

	for _, m := range mounts {
		t.Print(
			m.Path, fmtutil.PrettyNum(m.Listeners),
			timeutil.ShortDuration(m.Connected), m.ContentType,
		)
	}

	t.Separator()

	fmtc.NewLine()
}

// listClients prints info about clients (listeners) connected to given mount point
func listClients(mount string) {
	mount = formatMount(mount)
	listeners, err := client.ListClients(mount)

	if err != nil {
		printErrorExit(err.Error())
	}

	if len(listeners) == 0 {
		fmtc.Println("{y}No listeners found{!}")
		return
	}

	t := table.NewTable("id", "ip", "lag", "connected", "user-agent")
	t.SetAlignments(table.ALIGN_RIGHT, table.ALIGN_RIGHT, table.ALIGN_RIGHT, table.ALIGN_RIGHT)
	t.SetSizes(6, 14, 10, 9)

	fmtc.NewLine()

	for _, l := range listeners {
		t.Print(
			l.ID, l.IP, fmtutil.PrettySize(l.Lag),
			timeutil.ShortDuration(l.Connected),
			l.UserAgent,
		)
	}

	t.Separator()

	fmtc.NewLine()
}

// moveClients moves clients from one mount point to another
func moveClients(fromMount, toMount string) {
	fromMount = formatMount(fromMount)
	toMount = formatMount(toMount)

	err := client.MoveClients(fromMount, toMount)

	if err != nil {
		printErrorExit(err.Error())
	}

	fmtc.Printfn("{g}Clients successfully moved from %s to %s{!}", fromMount, toMount)
}

// updateMeta updates metadata for given mount point
func updateMeta(mount, artist, title string) {
	mount = formatMount(mount)

	err := client.UpdateMeta(mount, ic.TrackMeta{
		Artist: artist,
		Title:  title,
	})

	if err != nil {
		printErrorExit(err.Error())
	}

	fmtc.Printfn("{g}Metadata successfully updated for %s{!}", mount)
}

// killClient detaches client with given ID from the mount point
func killClient(mount, clientID string) {
	mount = formatMount(mount)
	id, err := strconv.Atoi(clientID)

	if err != nil {
		printErrorExit(err.Error())
	}

	err = client.KillClient(mount, id)

	if err != nil {
		printErrorExit(err.Error())
	}

	fmtc.Printfn("{g}Cliend %d successfully detached from %s{!}", id, mount)
}

// killSource detaches source from given mount point
func killSource(mount string) {
	mount = formatMount(mount)

	err := client.KillSource(mount)

	if err != nil {
		printErrorExit(err.Error())
	}

	fmtc.Printfn("{g}Source successfully detached from %s{!}", mount)
}

// printServerHeader prints header with icecast info
func printServerHeader(id string) {
	showSeparator(false)

	if id == "" {
		fmtc.Printfn(" {*}{#45}Icecast Server{!} on {*}%s{!}", options.GetS(OPT_HOST))
	} else {
		fmtc.Printfn(" {*}{#45}Icecast Server{!} on {*}%s{!} {s-}(%s){!}", options.GetS(OPT_HOST), id)
	}

	showSeparator(false)
}

// showSeparator prints separator
func showSeparator(shadow bool) {
	if shadow {
		fmtutil.SeparatorColorTag = "{s-}"
	} else {
		fmtutil.SeparatorColorTag = "{s}"
	}

	fmtutil.Separator(true)
}

// formatString formats string for stats info
func formatString(s string) string {
	if s == "" {
		return fmtc.Sprintf("{s-}—{!}")
	}

	return s
}

// formatMount formats mount name
func formatMount(mount string) string {
	if !strings.HasPrefix(mount, "/") {
		return "/" + mount
	}

	return mount
}

// checks command for required args num
func checkForRequiredArgs(args options.Arguments, required int) {
	if len(args) >= required+1 {
		return
	}

	printErrorExit(
		"Wrong number of arguments for %s command",
		args.Get(0).ToLower().String(),
	)
}

// printErrorExit prints error message to console and exit with error code
func printErrorExit(f string, a ...interface{}) {
	terminal.Error(f, a...)
	os.Exit(1)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// helpCmdStats shows help for "stats" command
func helpCmdStats() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Shows internal statistics kept by the Icecast server.\n")
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printfn("  {c*}%s{!} {y}%s{!}\n", APP, CMD_STATS)
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printfn("  %s %s", APP, CMD_STATS)
	fmtc.NewLine()
}

// helpCmdListMounts shows help for "list-mounts" command
func helpCmdListMounts() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Shows all the currently connected mountpoints.\n")
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printfn("  {c*}%s{!} {y}%s{!}\n", APP, CMD_LIST_MOUNTS)
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printfn("  %s %s", APP, CMD_LIST_MOUNTS)
	fmtc.NewLine()
}

// helpCmdListClients shows help for "list-clients" command
func helpCmdListClients() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Shows all the clients currently connected to a specific mountpoint.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printfn("  {c*}%s{!} {y}%s{!} {g}mount{!}", APP, CMD_LIST_CLIENTS)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}mount{!} - Mount name {s-}(with or without leading slash){!}")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printfn("  %s %s /source1.ogg", APP, CMD_LIST_CLIENTS)
	fmtc.Printfn("  %s %s source1.ogg", APP, CMD_LIST_CLIENTS)
	fmtc.NewLine()
}

// helpCmdMoveClients shows help for "move-clients" command
func helpCmdMoveClients() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  This command provides the ability to migrate currently connected listeners")
	fmtc.Println("  from one mountpoint to another.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printfn("  {c*}%s{!} {y}%s{!} {g}from-mount to-mount{!}", APP, CMD_MOVE_CLIENTS)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}from-mount{!} - Source mount name {s-}(with or without leading slash){!}")
	fmtc.Println("  {g}to-mount  {!} - Target mount name {s-}(with or without leading slash){!}")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printfn("  %s %s /source1.ogg /source2.ogg", APP, CMD_MOVE_CLIENTS)
	fmtc.Printfn("  %s %s source1.aac source2.aac", APP, CMD_MOVE_CLIENTS)
	fmtc.NewLine()
}

// helpCmdUpdateMeta shows help for "update-meta" command
func helpCmdUpdateMeta() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  This command provides the ability for either a source client or any external")
	fmtc.Println("  program to update the metadata information for a particular mountpoint.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printfn("  {c*}%s{!} {y}%s{!} {g}mount artist title{!}", APP, CMD_UPDATE_META)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}mount {!} - Mount name {s-}(with or without leading slash){!}")
	fmtc.Println("  {g}artist{!} - Track artist name")
	fmtc.Println("  {g}title {!} - Track title")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printfn("  %s %s \"Wretch 32\" \"Traktor (Brookes Brothers Remix)\"", APP, CMD_UPDATE_META)
	fmtc.NewLine()
}

// helpCmdKillClient shows help for "kill-client" command
func helpCmdKillClient() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Disconnects a specific listener of a currently connected mountpoint.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printfn("  {c*}%s{!} {y}%s{!} {g}mount{!}", APP, CMD_KILL_CLIENT)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}mount    {!} - Mount name {s-}(with or without leading slash){!}")
	fmtc.Println("  {g}client-id{!} - Client ID")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printfn("  %s %s /source1.ogg 457", APP, CMD_KILL_CLIENT)
	fmtc.Printfn("  %s %s source1.ogg 457", APP, CMD_KILL_CLIENT)
	fmtc.NewLine()
}

// helpCmdKillSource shows help for "kill-source" command
func helpCmdKillSource() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Disconnects a specific mountpoint from the server.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printfn("  {c*}%s{!} {y}%s{!} {g}mount{!", APP, CMD_KILL_SOURCE)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}mount{!} - Mount name {s-}(with or without leading slash){!}")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printfn("  %s %s /source1.ogg", APP, CMD_KILL_SOURCE)
	fmtc.Printfn("  %s %s source1.ogg", APP, CMD_KILL_SOURCE)
	fmtc.NewLine()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// printCompletion prints completion for given shell
func printCompletion() int {
	switch options.GetS(OPT_COMPLETION) {
	case "bash":
		fmt.Print(bash.Generate(genUsage(), APP))
	case "fish":
		fmt.Print(fish.Generate(genUsage(), APP))
	case "zsh":
		fmt.Print(zsh.Generate(genUsage(), optMap, APP))
	default:
		return 1
	}

	return 0
}

// printMan prints man page
func printMan() {
	fmt.Println(man.Generate(genUsage(), genAbout("")))
}

// genUsage generates usage info
func genUsage() *usage.Info {
	info := usage.NewInfo("icecli", "arguments…")

	info.AddCommand(CMD_STATS, "Show Icecast statistics")
	info.AddCommand(CMD_LIST_MOUNTS, "List mount points")
	info.AddCommand(CMD_LIST_CLIENTS, "List clients", "mount")
	info.AddCommand(CMD_MOVE_CLIENTS, "Move clients between mounts", "from-mount", "to-mount")
	info.AddCommand(CMD_UPDATE_META, "Update meta for mount", "mount", "artist", "title")
	info.AddCommand(CMD_KILL_CLIENT, "Kill client connection", "mount", "client-id")
	info.AddCommand(CMD_KILL_SOURCE, "Kill source connection", "mount")
	info.AddCommand(CMD_HELP, "Show detailed info about command usage", "command")

	info.AddOption(OPT_HOST, "URL of Icecast instance {s-}(default: http://127.0.0.1:8000){!}", "host")
	info.AddOption(OPT_USER, "Admin username {s-}(default: admin){!}", "username")
	info.AddOption(OPT_PASS, "Admin password {s-}(default: hackme){!}", "password")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	info.AddExample(
		CMD_STATS+" -H 127.0.0.1:10000",
		"Show stats for server on 127.0.0.1:10000",
	)

	info.AddExample(
		CMD_KILL_CLIENT+" -P mYsUpPaPaSs /stream3 361",
		"Detach client with ID 361 from /stream3",
	)

	info.AddExample(
		CMD_LIST_CLIENTS+" -H 127.0.0.1:10000 -U super_admin -P mYsUpPaPaSs /stream3",
		"List clients on /stream3",
	)

	return info
}

// genAbout generates info about version
func genAbout(gitRev string) *usage.About {
	about := &usage.About{
		App:     APP,
		Version: VER,
		Desc:    DESC,
		Year:    2009,
		Owner:   "ESSENTIAL KAOS",

		AppNameColorTag: colorTagApp,
		VersionColorTag: colorTagVer,
		DescSeparator:   "{s}—{!}",

		License:    "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
		BugTracker: "https://github.com/essentialkaos/icecli",
	}

	if gitRev != "" {
		about.Build = "git:" + gitRev
		about.UpdateChecker = usage.UpdateChecker{"essentialkaos/icecli", update.GitHubChecker}
	}

	return about
}
