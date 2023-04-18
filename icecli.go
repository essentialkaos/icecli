package main

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/fmtutil/table"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/timeutil"
	"github.com/essentialkaos/ek/v12/usage"
	"github.com/essentialkaos/ek/v12/usage/completion/bash"
	"github.com/essentialkaos/ek/v12/usage/completion/fish"
	"github.com/essentialkaos/ek/v12/usage/completion/zsh"
	"github.com/essentialkaos/ek/v12/usage/update"

	ic "github.com/essentialkaos/go-icecast/v2"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "icecli"
	DESC = "Icecast CLI"
	VER  = "1.0.2"
)

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

	OPT_COMPLETION = "completion"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var optMap = options.Map{
	OPT_HOST:     {Value: "http://127.0.0.1:8000", Alias: "url"},
	OPT_USER:     {Value: "admin", Alias: "login"},
	OPT_PASS:     {Value: "hackme", Alias: "pass"},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL, Alias: "u:usage"},
	OPT_VER:      {Type: options.BOOL, Alias: "ver"},

	OPT_COMPLETION: {},
}

var client *ic.API

// ////////////////////////////////////////////////////////////////////////////////// //

// main is main func
func main() {
	args, errs := options.Parse(optMap)

	if len(errs) != 0 {
		printError("Options parsing errors:")

		for _, err := range errs {
			printError("  %v", err)
		}

		os.Exit(1)
	}

	configureUI()

	if options.Has(OPT_COMPLETION) {
		genCompletion()
	}

	if options.GetB(OPT_VER) {
		showAbout()
		return
	}

	if options.GetB(OPT_HELP) || len(args) == 0 {
		showUsage()
		return
	}

	if args.Get(0).ToLower().String() == CMD_HELP {
		checkForRequiredArgs(args, 1)
		showHelp(args.Get(0).String())
	} else {
		execCommand(args)
	}
}

// configureUI configures user interface
func configureUI() {
	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}

	fmtutil.SeparatorSymbol = "–"
	fmtutil.SeparatorFullscreen = true
	fmtutil.SizeSeparator = " "
	table.SeparatorSymbol = "–"
	table.HeaderCapitalize = true
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

	switch args.Get(0).ToLower().String() {
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
		showUsage()
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
		showUsage()
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
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Sources", fmtutil.PrettyNum(stats.Stats.Sources))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Banned IPs", fmtutil.PrettyNum(stats.Stats.BannedIPs))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Clients", fmtutil.PrettyNum(stats.Stats.Clients))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Connections", fmtutil.PrettyNum(stats.Stats.Connections))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Listeners", fmtutil.PrettyNum(stats.Stats.Listeners))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Stats", fmtutil.PrettyNum(stats.Stats.Stats))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Client Connections", fmtutil.PrettyNum(stats.Stats.ClientConnections))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "File Connections", fmtutil.PrettyNum(stats.Stats.FileConnections))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Listener Connections", fmtutil.PrettyNum(stats.Stats.ListenerConnections))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Stats Connections", fmtutil.PrettyNum(stats.Stats.StatsConnections))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Source Client Connections", fmtutil.PrettyNum(stats.Stats.SourceClientConnections))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Source Relay Connections", fmtutil.PrettyNum(stats.Stats.SourceRelayConnections))
	fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Source Total Connections", fmtutil.PrettyNum(stats.Stats.SourceTotalConnections))
	fmtc.Printf(
		" {*}%-28s{!} {s}|{!} %s {s-}(%s){!}\n", "Stream Bytes Read",
		fmtutil.PrettyNum(stats.Stats.StreamBytesRead),
		fmtutil.PrettySize(stats.Stats.StreamBytesRead),
	)
	fmtc.Printf(
		" {*}%-28s{!} {s}|{!} %s {s-}(%s){!}\n", "Stream Bytes Sent",
		fmtutil.PrettyNum(stats.Stats.StreamBytesSent),
		fmtutil.PrettySize(stats.Stats.StreamBytesSent),
	)

	for path, source := range stats.Sources {
		showSeparator(false)
		fmtc.Printf(" {*y}%s{!} {s-}(online: %s){!}\n", path, timeutil.PrettyDuration(time.Since(source.StreamStarted)))
		showSeparator(false)
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Source IP", source.SourceIP)
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Name", formatString(source.Info.Name))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Genre", formatString(source.Genre))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Description", formatString(source.Info.Description))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Type", formatString(source.Info.Type))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "URL", formatString(source.Info.URL))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Listen URL", formatString(source.ListenURL))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "SubType", formatString(source.Info.SubType))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %t\n", "Public", source.Public)
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "User-Agent", formatString(source.UserAgent))
		showSeparator(true)
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Bitrate", fmtutil.PrettyNum(source.AudioInfo.Bitrate))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Channels", fmtutil.PrettyNum(source.AudioInfo.Channels))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s Hz\n", "SampleRate", fmtutil.PrettyNum(source.AudioInfo.SampleRate))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "CodecID", fmtutil.PrettyNum(source.AudioInfo.CodecID))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "RawInfo", formatString(source.AudioInfo.RawInfo))
		showSeparator(true)
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Artist", formatString(source.Track.Artist))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Title", formatString(source.Track.Title))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Artwork", formatString(source.Track.Artwork))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Metadata URL", formatString(source.Track.MetadataURL))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "RawInfo", formatString(source.Track.RawInfo))
		fmtc.Printf(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s ago){!}\n", "Metadata Updated",
			timeutil.Format(source.MetadataUpdated, "%Y/%m/%d %H:%M:%S"),
			timeutil.PrettyDuration(time.Since(source.MetadataUpdated)),
		)
		showSeparator(true)
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Listeners", fmtutil.PrettyNum(source.Stats.Listeners))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Listener Peak", fmtutil.PrettyNum(source.Stats.ListenerPeak))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Max Listeners", fmtutil.PrettyNum(source.Stats.MaxListeners))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Slow Listeners", fmtutil.PrettyNum(source.Stats.SlowListeners))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Listener Connections", fmtutil.PrettyNum(source.Stats.ListenerConnections))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Connected", fmtutil.PrettyNum(source.Stats.Connected))
		fmtc.Printf(" {*}%-28s{!} {s}|{!} %s\n", "Queue Size", fmtutil.PrettyNum(source.Stats.QueueSize))

		fmtc.Printf(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s/s){!}\n", "Incoming Bitrate",
			fmtutil.PrettyNum(source.Stats.IncomingBitrate),
			fmtutil.PrettySize(source.Stats.IncomingBitrate),
		)

		fmtc.Printf(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s/s){!}\n", "Outgoing Bitrate",
			fmtutil.PrettyNum(source.Stats.OutgoingBitrate),
			fmtutil.PrettySize(source.Stats.OutgoingBitrate),
		)

		fmtc.Printf(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s){!}\n", "Total Bytes Read",
			fmtutil.PrettyNum(source.Stats.TotalBytesRead),
			fmtutil.PrettySize(source.Stats.TotalBytesRead),
		)

		fmtc.Printf(
			" {*}%-28s{!} {s}|{!} %s {s-}(%s){!}\n", "Total Bytes Sent",
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

	fmtc.Printf("{g}Clients successfully moved from %s to %s{!}\n", fromMount, toMount)
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

	fmtc.Printf("{g}Metadata successfully updated for %s{!}\n", mount)
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

	fmtc.Printf("{g}Cliend %d successfully detached from %s{!}\n", id, mount)
}

// killSource detaches source from given mount point
func killSource(mount string) {
	mount = formatMount(mount)

	err := client.KillSource(mount)

	if err != nil {
		printErrorExit(err.Error())
	}

	fmtc.Printf("{g}Source successfully detached from %s{!}\n", mount)
}

// printServerHeader prints header with icecast info
func printServerHeader(id string) {
	showSeparator(false)

	if id == "" {
		fmtc.Printf(" {*}{#45}Icecast Server{!} on {*}%s{!}\n", options.GetS(OPT_HOST))
	} else {
		fmtc.Printf(" {*}{#45}Icecast Server{!} on {*}%s{!} {s-}(%s){!}\n", options.GetS(OPT_HOST), id)
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

// printError prints error message to console
func printError(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
}

// printErrorExit prints error message to console and exit with error code
func printErrorExit(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
	os.Exit(1)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func helpCmdStats() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Shows internal statistics kept by the Icecast server.\n")
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printf("  {c*}%s{!} {y}%s{!}\n\n", APP, CMD_STATS)
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printf("  %s %s\n", APP, CMD_STATS)
	fmtc.NewLine()
}

func helpCmdListMounts() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Shows all the currently connected mountpoints.\n")
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printf("  {c*}%s{!} {y}%s{!}\n\n", APP, CMD_LIST_MOUNTS)
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printf("  %s %s\n", APP, CMD_LIST_MOUNTS)
	fmtc.NewLine()
}

func helpCmdListClients() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Shows all the clients currently connected to a specific mountpoint.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printf("  {c*}%s{!} {y}%s{!} {g}mount{!}\n", APP, CMD_LIST_CLIENTS)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}mount{!} - Mount name {s-}(with or without leading slash){!}")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printf("  %s %s /source1.ogg \n", APP, CMD_LIST_CLIENTS)
	fmtc.Printf("  %s %s source1.ogg \n", APP, CMD_LIST_CLIENTS)
	fmtc.NewLine()
}

func helpCmdMoveClients() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  This command provides the ability to migrate currently connected listeners")
	fmtc.Println("  from one mountpoint to another.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printf("  {c*}%s{!} {y}%s{!} {g}from-mount to-mount{!}\n", APP, CMD_MOVE_CLIENTS)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}from-mount{!} - Source mount name {s-}(with or without leading slash){!}")
	fmtc.Println("  {g}to-mount  {!} - Target mount name {s-}(with or without leading slash){!}")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printf("  %s %s /source1.ogg /source2.ogg\n", APP, CMD_MOVE_CLIENTS)
	fmtc.Printf("  %s %s source1.aac source2.aac \n", APP, CMD_MOVE_CLIENTS)
	fmtc.NewLine()
}

func helpCmdUpdateMeta() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  This command provides the ability for either a source client or any external")
	fmtc.Println("  program to update the metadata information for a particular mountpoint.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printf("  {c*}%s{!} {y}%s{!} {g}mount artist title{!}\n", APP, CMD_UPDATE_META)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}mount {!} - Mount name {s-}(with or without leading slash){!}")
	fmtc.Println("  {g}artist{!} - Track artist name")
	fmtc.Println("  {g}title {!} - Track title")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printf("  %s %s \"Wretch 32\" \"Traktor (Brookes Brothers Remix)\"\n", APP, CMD_UPDATE_META)
	fmtc.NewLine()
}

func helpCmdKillClient() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Disconnects a specific listener of a currently connected mountpoint.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printf("  {c*}%s{!} {y}%s{!} {g}mount{!}\n", APP, CMD_KILL_CLIENT)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}mount    {!} - Mount name {s-}(with or without leading slash){!}")
	fmtc.Println("  {g}client-id{!} - Client ID")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printf("  %s %s /source1.ogg 457\n", APP, CMD_KILL_CLIENT)
	fmtc.Printf("  %s %s source1.ogg 457\n", APP, CMD_KILL_CLIENT)
	fmtc.NewLine()
}

func helpCmdKillSource() {
	fmtc.NewLine()
	fmtc.Println("{*}Description:{!}\n")
	fmtc.Println("  Disconnects a specific mountpoint from the server.")
	fmtc.NewLine()
	fmtc.Println("{*}Usage:{!}\n")
	fmtc.Printf("  {c*}%s{!} {y}%s{!} {g}mount{!}\n", APP, CMD_KILL_SOURCE)
	fmtc.NewLine()
	fmtc.Println("{*}Arguments:{!}\n")
	fmtc.Println("  {g}mount{!} - Mount name {s-}(with or without leading slash){!}")
	fmtc.NewLine()
	fmtc.Println("{*}Examples:{!}\n")
	fmtc.Printf("  %s %s /source1.ogg \n", APP, CMD_KILL_SOURCE)
	fmtc.Printf("  %s %s source1.ogg \n", APP, CMD_KILL_SOURCE)
	fmtc.NewLine()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// showUsage print usage info
func showUsage() {
	genUsage().Render()
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

// genCompletion generates completion for different shells
func genCompletion() {
	switch options.GetS(OPT_COMPLETION) {
	case "bash":
		fmt.Printf(bash.Generate(genUsage(), APP))
	case "fish":
		fmt.Printf(fish.Generate(genUsage(), APP))
	case "zsh":
		fmt.Printf(zsh.Generate(genUsage(), optMap, APP))
	default:
		os.Exit(1)
	}

	os.Exit(0)
}

// showAbout shows info about version
func showAbout() {
	about := &usage.About{
		App:           APP,
		Version:       VER,
		Desc:          DESC,
		Year:          2006,
		Owner:         "ESSENTIAL KAOS",
		License:       "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/icecli", update.GitHubChecker},
	}

	about.Render()
}
