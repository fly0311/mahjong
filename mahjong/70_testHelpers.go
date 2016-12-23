/*
mahjong: A computer-mediated Mah Jong game implemented in Go
Copyright (C) 2016 <code@0n0e.com>

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

package mahjong

import (
  "strings"
)

var gt *Game

func init() {
  gt = New()
  gt.InitializeForTesting()
}

func (g *Game) InitializeForTesting() {
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
}

func (gt *Game) TestHandMaker(unicodeTiles string) (PlayerHand, Tile) {
  portions := strings.Split(unicodeTiles, ";")
  
  var Hidden []Tile
  var Draw Tile
  drawProcessed := false
    
  // hidden
  for _, rune := range(portions[0]) {
    tileCharacter := string(rune)
    for i := 0; i < len(gt.Undealt); i++ {
      if tileCharacter == gt.Undealt[i].Ud {
        Hidden = append(Hidden, gt.Undealt[i])
        break
      }
    }
  }

  if len(portions[1]) > 0 {
    for _, rune := range(portions[1]) {
      
      if drawProcessed {
        break;
      }
      
      tileCharacter := string(rune)
      for i := 0; i < len(gt.Undealt); i++ {
        if tileCharacter == gt.Undealt[i].Ud {
          Draw = gt.Undealt[i]
          drawProcessed = true
          break
        }
      }
    }
  } else {
    Draw = EmptyTile
  }
  
  return PlayerHand{ Hidden: Hidden }, Draw
}
