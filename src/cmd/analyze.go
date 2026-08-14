package cmd

import (
	"Scout/src/demo"
	"Scout/src/utils"
	"fmt"
	"image"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyzes the demo specified.",
	Long:  `Analyzes the demo specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath, err := cmd.Flags().GetString("demofile")
		utils.CheckErrorm(err, "Failed to get demo file path.")

		cmd.Println("Analyzing: ", filePath)

		teams := getDemoTeams(filePath)
		teamKeys := utils.GetKeys(teams)

		promptText := []string{}
		for i, team := range teamKeys {
			teamMembers := teams[team]
			promptText = append(promptText, fmt.Sprintf("Team %d | Players: ", team))
			for j, teamMember := range teamMembers {
				text := promptText[i] + fmt.Sprintf("%s", *teamMember.Name)
				if j != len(teamMembers)-1 {
					text = text + ", "
				}
				slices.Replace(promptText, i, i+1, text)
			}
		}

		teamPrompt := promptui.Select{
			Label: "Select the team to analyze",
			Items: promptText,
		}

		teamIndex, _, err := teamPrompt.Run()
		utils.CheckErrorm(err, "Failed to select team.")
		fmt.Println("Ok.")

		playersOnTeam := []*msg.CCSUsrMsg_EndOfMatchAllPlayersData_PlayerData{}
		for i, _ := range teamKeys {
			if i == teamIndex {
				playersOnTeam = teams[teamKeys[i]]
			}
		}

		if len(playersOnTeam) == 0 {
			panic("No players on the selected team.")
		}

		startAnalysis(playersOnTeam, filePath)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.PersistentFlags().String("demofile", "d", "The file path of the (.dem) demo file.")
	analyzeCmd.MarkFlagRequired("demofile")
}

func getDemoTeams(demoPath string) map[int32][]*msg.CCSUsrMsg_EndOfMatchAllPlayersData_PlayerData {
	demoFile, err := os.Open(demoPath)
	utils.CheckErrorm(err, "Failed to read from the demo file.")

	parser := demoinfocs.NewParser(demoFile)
	teamList := map[int32][]*msg.CCSUsrMsg_EndOfMatchAllPlayersData_PlayerData{}

	parser.RegisterNetMessageHandler(func(message *msg.CCSUsrMsg_EndOfMatchAllPlayersData) {
		playerData := message.Allplayerdata
		for _, player := range playerData {
			if _, ok := teamList[*player.Teamnumber]; !ok {
				teamList[*player.Teamnumber] = []*msg.CCSUsrMsg_EndOfMatchAllPlayersData_PlayerData{}
			}
			teamList[*player.Teamnumber] = append(teamList[*player.Teamnumber], player)
		}
	})

	err = parser.ParseToEnd()
	utils.CheckError(err)

	return teamList
}

// Analysis stuff

var (
	pointsCT []demo.Position
	pointsT  []demo.Position
	mapData  demo.Map
	players  map[uint64]*msg.CCSUsrMsg_EndOfMatchAllPlayersData_PlayerData
	mapName  string
	radarImg image.Image
)

func onSmokeStart(event events.SmokeStart) {
	playerName := "Unknown"
	if event.Grenade != nil && event.Grenade.Owner != nil {
		playerName = event.Grenade.Owner.Name
	}

	if _, ok := players[event.Thrower.SteamID64]; ok {
		x, y := mapData.TranslateScale(event.Position.X, event.Position.Y)
		if event.Thrower.Team == 3 {
			pointsCT = append(pointsCT, demo.Position{PosX: x, PosY: y})
			fmt.Printf("CT: %s's threw smoke \n", playerName)
		} else {
			pointsT = append(pointsT, demo.Position{PosX: x, PosY: y})
			fmt.Printf("T: %s's threw smoke \n", playerName)
		}
	}
}

func startAnalysis(p []*msg.CCSUsrMsg_EndOfMatchAllPlayersData_PlayerData, demoPath string) {
	players = make(map[uint64]*msg.CCSUsrMsg_EndOfMatchAllPlayersData_PlayerData)
	for _, player := range p {
		players[*player.Xuid] = player
	}
	demoFile, err := os.Open(demoPath)
	utils.CheckErrorm(err, "Failed to read from the demo file.")

	parser := demoinfocs.NewParser(demoFile)

	parser.RegisterNetMessageHandler(func(message *msg.CSVCMsg_ServerInfo) {
		mapName = message.GetMapName()
		mapFile := demo.GetMapFile(mapName)
		radarImg = demo.GetImage(mapFile)
		mapData = demo.GetMapData(mapName)
		pointsCT = make([]demo.Position, 0)
		pointsT = make([]demo.Position, 0)
	})

	parser.RegisterEventHandler(onSmokeStart)

	err = parser.ParseToEnd()

	if err != nil {
		panic(err)
	}

	// CT
	outFilePath := "C:/Users/allison/Codespaces/CS/DemoFiles/CT_radar_with_smokes.png"
	fmt.Printf("Drawing %d CT smoke points to %s...\n", len(pointsCT), outFilePath)

	err = demo.DrawPointsAndSave(radarImg, pointsCT, outFilePath)
	if err != nil {
		fmt.Printf("Failed to save image: %v\n", err)
	} else {
		fmt.Println("Successfully saved radar image!")
	}

	// T
	outFilePath = "C:/Users/allison/Codespaces/CS/DemoFiles/T_radar_with_smokes.png"
	fmt.Printf("Drawing %d T smoke points to %s...\n", len(pointsT), outFilePath)

	err = demo.DrawPointsAndSave(radarImg, pointsT, outFilePath)
	if err != nil {
		fmt.Printf("Failed to save image: %v\n", err)
	} else {
		fmt.Println("Successfully saved radar image!")
	}
}
