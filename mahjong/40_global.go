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

// hold the global variables required for multiple instances (e.g., multiple games)
package mahjong

import(
  "log"
)

type Game struct {
  // # tileCollection
  // ordered stack of tiles
  Undealt TileCollection
  // shuffling can only occur once
  Shuffled bool
  // remaining tiles
  UndealtTileCount int
  // replacement draw location
  ReplacementPointer int
  // draw location
  DrawPointer int
  // draw locations set
  DrawLocationsSet bool
  // allocation start position
  AllocationStart int
  // discarded tiles
  Discard DiscardPile

  // # playerOps
  // player state
  Hands []PlayerHand

  // # stateMachineOps
  // current player
  CurrentPlayer int
  // dealer
  StartPlayer int
  
  // # throughout
  // output log
  OutputLog *log.Logger
}

func New() *Game {
  return &Game{}
}

// per game init
func (g *Game) Initialize(dealer int, computerPlayers []bool) {
  // # tileCollection
  g.UndealtTileCount = TilesInGame
  g.ReplacementPointer = -1 // to be initialized later
  g.DrawPointer = -1 // to be initialized later
  
  // allocate tile set
  g.Undealt = make([]Tile, TilesInGame, TilesInGame)
  // keep track of current tile
  p := 0 // position in tileCollection slice
  // create tiles for the standard suits and honor suit
  // TODO: clean magic numbers
  for i := 1; i <= 4; i++ {
    for j:= 1; j < 10; j++ {
      for k:= 0; k < 4; k++ {
        t := Tile { 
          Suit: i, 
          Value: j, 
          Id: (i-1)*36+(j-1)*4+k+1, 
          Ud: UnicodeDisplay[i-1][j] }
        g.Undealt[p] = t
        p++
      }
      if i == 4 && j == 7 { // bail out for honor suit
        break
      }
    }
  }

  // create tiles for the bonus suit
  for i := 5; i <= 5; i++ {
    for j:= 1; j <= 8; j++ {
      t := Tile{ 
          Suit: i, 
          Value: j, 
          Id: 3*36+7*4+j,
          Ud: UnicodeDisplay[i-1][j] }
      g.Undealt[p] = t
      p++
    }
  }
  
  // # playerOps
  g.Hands = make([]PlayerHand, PlayersInGame, PlayersInGame)
  for i := 0; i < PlayersInGame; i++ {
    g.Hands[i].Hidden = make([]Tile, 14, 14)
    g.Hands[i].Revealed = make([]Tile, 8, 8)
    g.Hands[i].Player = i
    g.Hands[i].ComputerPlayer = computerPlayers[i]
  }

  // # stateMachineOps
  if dealer != -1 {
    g.CurrentPlayer = dealer
    g.StartPlayer = dealer
  }
  
  // # other initialization
  // shuffle tiles
  err := g.Shuffle(DeterministicRand)
  if err != nil {
    log.Fatal(err)
  }
  
  // output undealt tiles
  if VerboseDebug {
    g.OutputUndealtTiles();
  }
    
  // simulate dice roll to set deal locations
  diceRoll := RollOneDice(DeterministicRand)+RollOneDice(DeterministicRand)+RollOneDice(DeterministicRand)
  g.OutputLog.Println("dice roll:", diceRoll)
  
  err = g.SetDealLocations(diceRoll)
  if err != nil {
    log.Fatal(err)
  }
  
  if dealer == -1 {
    g.CurrentPlayer = (diceRoll-1) % 4
    g.StartPlayer = g.CurrentPlayer
  }
  
  // deal initial set of tiles
  g.InitialDeal()

  // replace special tiles
  g.InitialHandleSpecialTiles()
  
  // dump hands
  if VerboseDebug {
    for i := 0; i < 4; i++ {
      g.Hands[i].OutputHand(false, true) // do not show hidden
      //mahjong.Hands[i].OutputHand(true, true) // show hidden
    }    
  }
}
