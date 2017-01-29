/*
mahjong: A computer-mediated Mah Jong game implemented in Go
Copyright (C) 2016-7 <code@0n0e.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

// handle state machine operations
package mahjong

import(
  "fmt"
  "strconv"
  "strings"
  "log"
)

// state unit
type StateUnit struct {
  Player int
  State string
  Phase string
}

// end states
var EndStates map[string]bool

// initialize end states
func init() {  
  EndStates = make(map[string]bool)
  EndStates["WinGameP0"] = true
  EndStates["WinGameP1"] = true
  EndStates["WinGameP2"] = true
  EndStates["WinGameP3"] = true
  EndStates["DrawGame"] = true
}

func (g *Game) handToPlayer(newPlayer int) {
  if g.CurrentPlayer != newPlayer {
    // clear screen + request to be handed over to new player
    fmt.Printf("\u001b[2J")
    fmt.Printf("Next action is to be completed by player %d. Please have them drop by.\n", newPlayer)
    var input string
    fmt.Scanln(&input)
    g.CurrentPlayer = newPlayer
    fmt.Printf("\u001b[2J")
  }
}


// begin game
func (g *Game) BeginGame()(bool, int) {
  stateObj := StateUnit { Player: g.CurrentPlayer, State: "HaveWin", Phase: "DrawProcessing" }
  
  g.OutputLog.Println("gameplay begins with player", g.CurrentPlayer)
  
  if !g.Hands[g.CurrentPlayer].ComputerPlayer {
    // clear screen
    // TODO: deduplicate code with handToPlayer
    fmt.Printf("\u001b[2J")
    fmt.Printf("Game is to be started by player %d. Please have them drop by.\n", g.CurrentPlayer)
    var input string
    fmt.Scanln(&input)
    fmt.Printf("\u001b[2J")
  }
  
  var nextState StateUnit
  
  // start running
  for {
    nextState = g.processState(stateObj)
    
    if EndStates[nextState.State] {
      fmt.Printf("Game ended: %v\n", nextState.State)    
      g.OutputLog.Println("gameplay ends with outcome", nextState.State)
      
      g.OutputDiscardedTiles()
      g.Hands[0].OutputHand(true,true)
      g.Hands[1].OutputHand(true,true)
      g.Hands[2].OutputHand(true,true)
      g.Hands[3].OutputHand(true,true)
      break;
    } else {
      stateObj = nextState
    }
  }
  
  switch nextState.State {
    case "WinGameP0": 
      return true, 0
    case "WinGameP1": 
      return true, 1
    case "WinGameP2": 
      return true, 2
    case "WinGameP3": 
      return true, 3
    case "DrawGame": 
      return false, g.StartPlayer
  }
  return false, g.StartPlayer
}

// show game state
func (g *Game) ShowGameState(reveal bool, player int, showLatestTile bool) {
  // clear screen
  fmt.Printf("\u001b[2J")
  
  g.OutputDiscardedTiles()
  fmt.Printf("%d new tiles remain\n\n", g.UndealtTileCount)
  
  g.Hands[(player+3)%4].OutputHand(false,true)
  g.Hands[(player+2)%4].OutputHand(false,true)
  g.Hands[(player+1)%4].OutputHand(false,true)

  g.Hands[player].OutputHand(true,true)
  
  if showLatestTile && g.Hands[player].LastNewTile != EmptyTile {
    fmt.Printf("P%d-N: %v\n", player, g.Hands[player].LastNewTile.Ud)
  }
  
}

// process state to get next state
func (g *Game) processState(curState StateUnit) StateUnit {
  if curState.State == "HaveWin" && curState.Phase == "DrawProcessing" {
    // does the player have a winning hand?
    if g.Hands[curState.Player].HaveWin(EmptyTile, "draw") {
      var input string
      
      if !g.Hands[curState.Player].ComputerPlayer {
        g.handToPlayer(curState.Player)
        g.ShowGameState(false, curState.Player, true)
          
        fmt.Printf("Player %d: You appear to have a win. Do you take it? (y/n) [y]\n", curState.Player)
        fmt.Scanln(&input)
      } else {
        input = g.Hands[curState.Player].TakeWin(g.Discard, false, g.Hands)
      }
      
      if input == "" || input == "y" {
        g.OutputLog.Printf("player %d chose to take the win\n", curState.Player)
        
        return StateUnit { Player: curState.Player, State: "WinGameP"+strconv.Itoa(curState.Player), Phase: "DrawProcessing" }
      }
    }

    if VerboseDebug {
      fmt.Printf("[vd] Player %d: No win at this time; moving on to kong check.\n", curState.Player)
    }
    return StateUnit { Player: curState.Player, State: "HaveKong", Phase: "DrawProcessing" }
  } else if curState.State == "HaveKong" && curState.Phase == "DrawProcessing" {
    
    if kongResult, kongOptions := g.Hands[curState.Player].HaveKong(EmptyTile, "draw"); kongResult {
    
      var input string
      
      if !g.Hands[curState.Player].ComputerPlayer {
        g.handToPlayer(curState.Player)
        g.ShowGameState(false, curState.Player, true)
          
        fmt.Printf("Player %d: You appear to have at least one set of four. Do you take it, if so, which? (#) [n]\n", curState.Player)
        for i := 0; i < len(kongOptions); i++ {
          fmt.Printf("Option %d: Set of 4: %v\n", i, kongOptions[i].Tiles)
        }
        fmt.Scanln(&input)
      } else {
        input = g.Hands[curState.Player].TakeKong(g.Discard, true, g.Hands)
      }
      
      selection, _ := strconv.Atoi(input)
      
      if input != "n" && (selection >= 0 || selection < len(kongOptions)) {
        counter := 0
        for i := 0; i < 14; i++ {
          if g.Hands[curState.Player].Hidden[i] != EmptyTile && strings.Contains(kongOptions[selection].Tiles, g.Hands[curState.Player].Hidden[i].Ud) {
            g.Hands[curState.Player].Hidden[i] = EmptyTile
            counter++
          }
        }
        
        //g.OutputLog.Printf("kong? %d\n", counter)
        
        if counter > 2 {
          // move set away
          g.Hands[curState.Player].RevealedTileSets = append(g.Hands[curState.Player].RevealedTileSets, kongOptions[selection])
          g.Hands[curState.Player].RevealedSets++
        } else {
          //g.OutputLog.Printf("kong needs update\n")
          // update set
          for i := 0; i < g.Hands[curState.Player].RevealedSets; i++ {
            if g.Hands[curState.Player].RevealedTileSets[i].Kind == "triple" && strings.Contains(kongOptions[selection].Tiles, g.Hands[curState.Player].RevealedTileSets[i].Tiles) {
              // update
              g.Hands[curState.Player].RevealedTileSets[i] = kongOptions[selection]
              //g.OutputLog.Printf("kong updated\n")
              break
            }
          }
        }
        
        g.OutputLog.Printf("player %d reveals kong comprising %s\n", curState.Player, kongOptions[selection].Tiles)
        
        return StateUnit { Player: curState.Player, State: "DrawReplacementTile", Phase: "DrawProcessing" }

      }
      
    }
    
    if VerboseDebug {
      fmt.Printf("[vd] Player %d: No kong at this time; moving on to discard processing.\n", curState.Player)
    }
    return StateUnit { Player: curState.Player, State: "Discard", Phase: "DrawProcessing" }
  } else if curState.State == "DrawReplacementTile" {
    newTile, err := g.GetNewTile(&g.ReplacementPointer, true) 
    
    if err != nil {
      return StateUnit { Player: curState.Player, State: "DrawGame", Phase: "DrawProcessing" }
    }
    
    err = g.Hands[curState.Player].Receive(newTile)
    if err != nil {
      log.Fatal(err)
    }
    
    if VerboseDebug {
      fmt.Printf("[vd] Player %d drew as replacement %v\n", curState.Player, newTile)
    }
    
    g.Hands[curState.Player].LastNewTile = newTile
    
    return StateUnit { Player: curState.Player, State: "HandleSpecialTile", Phase: "DrawProcessing" }
  } else if curState.State == "HandleSpecialTile" {
    for g.Hands[curState.Player].HasUnreplacedSpecialTile() {
      specialTile, err := g.Hands[curState.Player].GetFirstSpecialTile()
      if err != nil {
        log.Fatal(err)
      }
      
      err = g.Hands[curState.Player].RevealSpecialTile(specialTile)
      if err != nil {
        log.Fatal(err)
      }
      g.OutputLog.Printf("player %d reveals special tile %s\n", curState.Player, specialTile.Ud)
      
      newTile, err := g.GetNewTile(&g.ReplacementPointer, true)
      if err != nil {
        return StateUnit { Player: curState.Player, State: "DrawGame", Phase: "DrawProcessing" }
      }
            
      g.Hands[curState.Player].Receive(newTile)
      if err != nil {
        log.Fatal(err)
      }
      
      if VerboseDebug {
        fmt.Printf("[vd] Player %d had special tile %v, which was replaced with tile %v\n", curState.Player, specialTile, newTile)
      }
      
      g.Hands[curState.Player].LastNewTile = newTile
    }
    return StateUnit { Player: curState.Player, State: "HaveWin", Phase: "DrawProcessing" }
  } else if curState.State == "DrawTile" {
    newTile, err := g.GetNewTile(&g.DrawPointer, false)
    if err != nil {
      return StateUnit { Player: curState.Player, State: "DrawGame", Phase: "DrawProcessing" }
    }
    
    g.OutputLog.Printf("player %d drew a tile\n", curState.Player)
    
    err = g.Hands[curState.Player].Receive(newTile)
    if err != nil {
      log.Fatal(err)
    }    
    
    if VerboseDebug {
      fmt.Printf("[vd] Player %d drew tile %v\n", curState.Player, newTile)
    }
    
    g.Hands[curState.Player].LastNewTile = newTile
    
    return StateUnit { Player: curState.Player, State: "HandleSpecialTile", Phase: "DrawProcessing" }
  } else if curState.State == "Discard" {
    var input string
    var discardSuggestion = g.Hands[curState.Player].Discard(g.Discard, false, g.Hands)
      
    if !g.Hands[curState.Player].ComputerPlayer {
      g.handToPlayer(curState.Player)
      g.ShowGameState(false, curState.Player, true)
  
      helperLine := ""
      for i := 0; i < 14; i++ {
        if g.Hands[curState.Player].Hidden[i] != EmptyTile {
          helperLine += fmt.Sprintf("(%s%d)", g.Hands[curState.Player].Hidden[i].Ud, i)
        }
      }
      fmt.Printf("%s\n", helperLine)
      
      // which tile does the player wish to discard?
      
      fmt.Printf("Player %d: What do you want to discard? # [%s]\n", curState.Player, discardSuggestion)

      fmt.Scanln(&input)
    } else {
      input = discardSuggestion
    }
    
    if len(input) == 0 {
      input = discardSuggestion
    }
    
    selection, _ := strconv.Atoi(input)
    
    if selection < 0 || selection > 13 || g.Hands[curState.Player].Hidden[selection] == EmptyTile {
      selection = 0
    }
    
    newDiscard := DiscardedTile { Player: curState.Player, Item: g.Hands[curState.Player].Hidden[selection] }
    
    g.Hands[curState.Player].Hidden[selection] = EmptyTile
   
    g.Discard = append(g.Discard, newDiscard)
    
    g.OutputLog.Printf("player %d discards tile %s\n", curState.Player, newDiscard.Item.Ud)
    
    g.Hands[curState.Player].LastNewTile = EmptyTile
    
    if VerboseDebug {
      fmt.Printf("[vd] Player %d chose to discard %v\n", curState.Player, newDiscard.Item.Ud)
    }

    return StateUnit { Player: (curState.Player+1) % 4, State: "HaveWin", Phase: "DiscardProcessing" }
  } else if curState.State == "HaveWin" && curState.Phase == "DiscardProcessing" {
    relationship := ""
    if (g.Discard[len(g.Discard)-1].Player + 1) % 4 == curState.Player {
      relationship = "previous"
    } else {
      relationship = "other"
    }
    
    // does the player have a winning hand?
    if g.Hands[curState.Player].HaveWin(g.Discard[len(g.Discard)-1].Item, relationship) {
      var input string
      
      if !g.Hands[curState.Player].ComputerPlayer {
        g.handToPlayer(curState.Player)
        g.ShowGameState(false, curState.Player, true)
        
        fmt.Printf("Player %d: You appear to have a win if you add in the discarded tile %v. Do you take it? (y/n) [y]\n", curState.Player, g.Discard[len(g.Discard)-1].Item)

        fmt.Scanln(&input)
      } else {
        input = g.Hands[curState.Player].TakeWin(g.Discard, true, g.Hands)
      }
      
      if input == "" || input == "y" {
        g.OutputLog.Printf("player %d chose to take the win with use of the discarded tile\n", curState.Player)
        
        return StateUnit { Player: curState.Player, State: "WinGameP"+strconv.Itoa(curState.Player), Phase: "DiscardProcessing" }
      }
    }

    if VerboseDebug {
      fmt.Printf("[vd] Player %d: No win at this time; moving on to next player.\n", curState.Player)
    }
    
    if (curState.Player + 1) % 4 == g.Discard[len(g.Discard)-1].Player {
      return StateUnit { Player: (curState.Player + 2) % 4, State: "HaveKong", Phase: "DiscardProcessing" }
    } else {
      return StateUnit { Player: (curState.Player + 1) % 4, State: "HaveWin", Phase: "DiscardProcessing" }
    }
  } else if curState.State == "HaveKong" && curState.Phase == "DiscardProcessing" {
    relationship := ""
    if (g.Discard[len(g.Discard)-1].Player + 1) % 4 == curState.Player {
      relationship = "previous"
    } else {
      relationship = "other"
    }
    
    if kongResult, kongOptions := g.Hands[curState.Player].HaveKong(g.Discard[len(g.Discard)-1].Item, relationship); kongResult {
      var input string
      
      if !g.Hands[curState.Player].ComputerPlayer {
        g.handToPlayer(curState.Player)
        g.ShowGameState(false, curState.Player, true)
        
        fmt.Printf("Player %d: You appear to have one set of four. Do you take it? (#) [n]\n", curState.Player)
        for i := 0; i < len(kongOptions); i++ {
          fmt.Printf("Option %d: Set of 4: %v\n", i, kongOptions[i].Tiles)
        }
      } else {
        input = g.Hands[curState.Player].TakeKong(g.Discard, true, g.Hands)
      }
      
      fmt.Scanln(&input)
      
      selection, _ := strconv.Atoi(input)
      
      if input != "n" && (selection >= 0 || selection < len(kongOptions)) {
        counter := 0
        for i := 0; i < 14; i++ {
          if g.Hands[curState.Player].Hidden[i] != EmptyTile && strings.Contains(kongOptions[selection].Tiles, g.Hands[curState.Player].Hidden[i].Ud) {
            g.Hands[curState.Player].Hidden[i] = EmptyTile
            counter++
          }
        }
        
        //g.OutputLog.Printf("kong? %d\n", counter)
        
        if counter > 2 {
          // can only move set away on discard
          g.Hands[curState.Player].RevealedTileSets = append(g.Hands[curState.Player].RevealedTileSets, kongOptions[selection])
          g.Hands[curState.Player].RevealedSets++
        } else {
          //g.OutputLog.Printf("kong needs update\n")
          // update set
          for i := 0; i < g.Hands[curState.Player].RevealedSets; i++ {
            if g.Hands[curState.Player].RevealedTileSets[i].Kind == "triple" && strings.Contains(kongOptions[selection].Tiles, g.Hands[curState.Player].RevealedTileSets[i].Tiles) {
              // update
              g.Hands[curState.Player].RevealedTileSets[i] = kongOptions[selection]
              //g.OutputLog.Printf("kong updated\n")
              break
            }
          }
        }
        
        g.OutputLog.Printf("player %d reveals kong comprising %s\n", curState.Player, kongOptions[selection].Tiles)
        
        return StateUnit { Player: curState.Player, State: "DrawReplacementTile", Phase: "DrawProcessing" }
      }
      
    }
    
    if VerboseDebug {
      fmt.Printf("[vd] Player %d: No kong at this time; moving on to next player.\n", curState.Player)
    }

    if (curState.Player + 1) % 4 == g.Discard[len(g.Discard)-1].Player {
      return StateUnit { Player: (curState.Player + 2) % 4, State: "HavePong", Phase: "DiscardProcessing" }
    } else {
      return StateUnit { Player: (curState.Player + 1) % 4, State: "HaveKong", Phase: "DiscardProcessing" }
    }
  } else if curState.State == "HavePong" && curState.Phase == "DiscardProcessing" {
    relationship := ""
    if (g.Discard[len(g.Discard)-1].Player + 1) % 4 == curState.Player {
      relationship = "previous"
    } else {
      relationship = "other"
    }
    
    if pongResult, pong := g.Hands[curState.Player].HavePong(g.Discard[len(g.Discard)-1].Item, relationship); pongResult && pong == g.Discard[len(g.Discard)-1].Item.Ud {
      var input string
      
      if !g.Hands[curState.Player].ComputerPlayer {
        g.handToPlayer(curState.Player)
        g.ShowGameState(false, curState.Player, true)
        
        fmt.Printf("Player %d: You can have a pong of %v with the most recent discard. Do you take it? (y/n) [y]\n", curState.Player, pong)
        fmt.Scanln(&input)
      } else {
        input = g.Hands[curState.Player].TakePong(g.Discard, true, g.Hands)
      }
      
      if input == "" || input == "y" {
        // move set away
        pongSet := TileSet{ Kind: "triple", Tiles: pong+pong+pong }
        
        g.Hands[curState.Player].RevealedTileSets = append(g.Hands[curState.Player].RevealedTileSets, pongSet)

        g.Hands[curState.Player].RevealedSets++
        counter := 0
        for i := 0; i < 14 && counter < 2; i++ {
          if g.Hands[curState.Player].Hidden[i].Ud == pong {
            g.Hands[curState.Player].Hidden[i] = EmptyTile
            counter++
          }
        }
                
        // remove last discard
        g.Discard = g.Discard[:len(g.Discard)-1]

        g.OutputLog.Printf("player %d reveals pong comprising %s\n", curState.Player, pongSet.Tiles)

        return StateUnit { Player: curState.Player, State: "Discard", Phase: "DrawProcessing" }
      }
    }
    
    if VerboseDebug {
      fmt.Printf("[vd] Player %d: No pong at this time; moving on to next player.\n", curState.Player)
    }
    
    if (curState.Player + 1) % 4 == g.Discard[len(g.Discard)-1].Player {
      return StateUnit { Player: (curState.Player + 2) % 4, State: "HaveSeq", Phase: "DiscardProcessing" }
    } else {
      return StateUnit { Player: (curState.Player + 1) % 4, State: "HavePong", Phase: "DiscardProcessing" }
    }
  } else if curState.State == "HaveSeq" && curState.Phase == "DiscardProcessing" {
    relationship := ""
    if (g.Discard[len(g.Discard)-1].Player + 1) % 4 == curState.Player {
      relationship = "previous"
    } else {
      relationship = "other"
    }
    
    if seqResult, seqOptions := g.Hands[curState.Player].HaveSeq(g.Discard[len(g.Discard)-1].Item, relationship); seqResult {
      var input string
      
      if !g.Hands[curState.Player].ComputerPlayer {
        g.handToPlayer(curState.Player)
        g.ShowGameState(false, curState.Player, true)
      
        fmt.Printf("Player %d: Using the most recent discard %v, you can form the following sequence(s): Do you take it, if so, which? (#) [n]\n", curState.Player, g.Discard[len(g.Discard)-1].Item.Ud)
      
        for i := 0; i < len(seqOptions); i++ {
          fmt.Printf("Option %d: Set of 3: %v\n", i, seqOptions[i].Tiles)
        }      
        fmt.Scanln(&input)
      } else {
        input = g.Hands[curState.Player].TakeSeq(g.Discard, true, g.Hands, seqOptions)
      }
      
      selection, _ := strconv.Atoi(input)
    
      if input != "n" && (selection >= 0 || selection < len(seqOptions)) {        
        // move set away
        g.Hands[curState.Player].RevealedTileSets = append(g.Hands[curState.Player].RevealedTileSets, seqOptions[selection])
        g.Hands[curState.Player].RevealedSets++
        
        // remove tiles from hand
        for _, runeValue := range seqOptions[selection].Tiles {
          counter := 0
          for i := 0; i < 14 && counter < 1; i++ {
            if g.Hands[curState.Player].Hidden[i].Ud == string(runeValue) && string(runeValue) != g.Discard[len(g.Discard)-1].Item.Ud {
              g.Hands[curState.Player].Hidden[i] = EmptyTile
              counter++
            }
          }
        }
        
        // remove last discard
        g.Discard = g.Discard[:len(g.Discard)-1]
        
        g.OutputLog.Printf("player %d reveals seq comprising %s\n", curState.Player, seqOptions[selection].Tiles)
        
        return StateUnit { Player: curState.Player, State: "Discard", Phase: "DrawProcessing" }
      }
    }
    
    if VerboseDebug {
      fmt.Printf("[vd] Player %d: No seq at this time; moving on to next player.\n", curState.Player)
    }
    
    return StateUnit { Player: curState.Player, State: "DrawTile", Phase: "DrawProcessing" }
  } else {
    fmt.Printf("Unknown state: %v", curState)
    // default outcome for a missing state
    return StateUnit { Player: curState.Player, State: "DrawGame", Phase: "DrawProcessing" } 
  }
}

