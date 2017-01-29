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

// handle player operations
package mahjong

import(
  "fmt"
  "log"
  "strings"
  "strconv"
  insecureRand "math/rand"
//  "errors"
)

// # player hand
const (
  PlayersInGame = 4
)

// tiles that form a set
type TileSet struct {
// TODO: populate UnderlyingTiles for audit checks
  Kind string
  Tiles string
  UnderlyingTiles []Tile
}

// hand of a player
type PlayerHand struct {
  Hidden []Tile
  Revealed []Tile
  RevealedSets int
  RevealedTileSets []TileSet
  Player int
  LastNewTile Tile
  ComputerPlayer bool
}

// max value for each suit
var MaxTileIndex []int

func init() {
  MaxTileIndex = make([]int, 4, 4)
  MaxTileIndex[0] = 9
  MaxTileIndex[1] = 9
  MaxTileIndex[2] = 9
  MaxTileIndex[3] = 7
}

// output hand
func (h PlayerHand) OutputHand(showHidden bool, tileOnly bool) {
  h.Sort()
  // public
  //fmt.Printf("Player %d's hand\n", h.Player)

  specialTileLine := ""
  
  specialTileLine += fmt.Sprintf("P%d-R: ", h.Player)
  for i := 0; i < 8; i++ {
    if h.Revealed[i] != EmptyTile {
      if tileOnly {
        specialTileLine += fmt.Sprintf("%v", h.Revealed[i].UdString())
      } else {
        specialTileLine += fmt.Sprintf("%v", h.Revealed[i])
      }
    }
  }
  fmt.Println(specialTileLine)
  
  if h.RevealedSets > 0 {
    fmt.Printf("P%d-R: ", h.Player)
    for i := 0; i < h.RevealedSets; i++ {
      if i > 0 {
        fmt.Printf(", ")
      }
      fmt.Printf("%v", h.RevealedTileSets[i].Tiles)
    }
    fmt.Println()
  }
  
  // private
  if showHidden {
    fmt.Printf("P%d-H: ", h.Player)
    for i := 0; i < 14; i++ {
      if h.Hidden[i] != EmptyTile {
        if tileOnly {
          fmt.Printf("%v", h.Hidden[i].UdString())
        } else {
          fmt.Printf("%v", h.Hidden[i])
        }
      }
    }
    fmt.Println()
  }
  
  fmt.Println()
}

// sort hand for readability/keep clear the last tile for new tiles
func (h PlayerHand) Sort() {
  // TODO: sort more efficiently or just sort on tile serial?
  for i := 0; i < 14; i++ {
    baseTile := EmptyTile
    baseItem := i
    
    for j := i; j < 14; j++ {
      if baseTile == EmptyTile && h.Hidden[j] != EmptyTile {
        baseTile = h.Hidden[j]
        baseItem = j
      }
      if baseTile != EmptyTile && h.Hidden[j] != EmptyTile && (h.Hidden[j].Suit <= baseTile.Suit || h.Hidden[j].Value <= baseTile.Value) {
        if h.Hidden[j].Suit == baseTile.Suit && h.Hidden[j].Value <= baseTile.Value {
          baseTile = h.Hidden[j]
          baseItem = j
        } else if h.Hidden[j].Suit != baseTile.Suit && h.Hidden[j].Suit <= baseTile.Suit {
          baseTile = h.Hidden[j]
          baseItem = j
        }
      }
    }
    h.Hidden[i], h.Hidden[baseItem] = h.Hidden[baseItem], h.Hidden[i] 
  }
}

// add tile to hand
func (h PlayerHand) Receive(t Tile) error {
  // sort if needed
  if h.Hidden[13] != EmptyTile {
    h.Sort()
  }
  
  if h.Hidden[13] == EmptyTile {
    h.Hidden[13] = t
  } else {
    return fmt.Errorf("tile %v could not be placed in hand as the last position was occupied by %v", t, h.Hidden[13])
  }
  h.Sort()
  return nil
}

// determine if hand has at least one special tile
func (h PlayerHand) HasUnreplacedSpecialTile() bool {
  for i := 0; i < 14; i++ {
    if h.Hidden[i].IsSpecial() {
      return true
    }
  }
  return false
}

// return the first special tile
func (h PlayerHand) GetFirstSpecialTile() (Tile, error) {
  for i := 0; i < 14; i++ {
    if h.Hidden[i].IsSpecial() {
      foundTile := h.Hidden[i]
      h.Hidden[i] = EmptyTile
      return foundTile, nil
    }
  }
  return EmptyTile, fmt.Errorf("Player %d's hand does not have any special tiles", h.Player)
}

// reveal special tile
func (h PlayerHand) RevealSpecialTile (t Tile) error {
  for k:= 0; k < 8; k++ {
    if h.Revealed[k] == EmptyTile {
      h.Revealed[k] = t
      
      return nil
    }
  }
  return fmt.Errorf("Player %d's special tile store was full and could not accommodate a special tile", h.Player)
}

// compute tile counts for hidden portion of hand and optional additional tile
func (h PlayerHand) CountHiddenTiles(consider Tile) ([][]int, []int, []int) {
  // counts
  tileCounts := make([][]int, 4, 4)
  for i := 0; i < 4; i++ {
    tileCounts[i] = make([]int, 10, 10)
  } // allocating an extra two positions for the honor suit

  // number of tiles in each suit
  tileCountsSum := make([]int, 4, 4)
  
  // sum of values, applicable only to the first three suits
  tileValuesSum := make([]int, 4, 4)
  
  // begin counting tiles
  if consider != EmptyTile {
    tileCounts[consider.Suit-1][consider.Value]++
    tileCountsSum[consider.Suit-1]++
    tileValuesSum[consider.Suit-1] += consider.Value
  }
  
  for i:= 0; i < len(h.Hidden); i++ {
    if h.Hidden[i] != EmptyTile {
      tmpTile := h.Hidden[i]
      tileCounts[tmpTile.Suit-1][tmpTile.Value]++
      tileCountsSum[tmpTile.Suit-1]++
      tileValuesSum[tmpTile.Suit-1] += tmpTile.Value
    }
  }
  
  return tileCounts, tileCountsSum, tileValuesSum
}

// compute tile counts for discards, optionally considering the last tile
func (d DiscardPile) CountDiscardTiles(ignoreLast bool) ([][]int, []int, []int) {
  // counts
  tileCounts := make([][]int, 4, 4)
  for i := 0; i < 4; i++ {
    tileCounts[i] = make([]int, 10, 10)
  } // allocating an extra two positions for the honor suit

  // number of tiles in each suit
  tileCountsSum := make([]int, 4, 4)
  
  // sum of values, applicable only to the first three suits
  tileValuesSum := make([]int, 4, 4)
  
  // begin counting tiles
  tileCount := len(d)
  if ignoreLast && tileCount > 0 {
    tileCount--
  }

  for i:= 0; i < tileCount; i++ {
    if d[i].Item != EmptyTile {
      tmpTile := d[i].Item
      tileCounts[tmpTile.Suit-1][tmpTile.Value]++
      tileCountsSum[tmpTile.Suit-1]++
      tileValuesSum[tmpTile.Suit-1] += tmpTile.Value
    }
  }
  
  return tileCounts, tileCountsSum, tileValuesSum
}

// compute tile counts for public portion of hand
func (h PlayerHand) CountPublicSetTiles() ([][]int, []int, []int) {
  // counts
  tileCounts := make([][]int, 4, 4)
  for i := 0; i < 4; i++ {
    tileCounts[i] = make([]int, 10, 10)
  } // allocating an extra two positions for the honor suit

  // number of tiles in each suit
  tileCountsSum := make([]int, 4, 4)
  
  // sum of values, applicable only to the first three suits
  tileValuesSum := make([]int, 4, 4)
  
  // TODO: make more efficient when underlying tiles are stored
  // begin counting tiles
  for i:= 0; i < len(h.RevealedTileSets); i++ {
    // map each tile
    for _, rune := range h.RevealedTileSets[i].Tiles {
      for m := 0; m < 4; m++ {
        for n, rud := range UnicodeDisplay[m] {
          if rud == string(rune) {
            tileCounts[m][n]++
            tileCountsSum[m]++
            tileValuesSum[m] += n
          }
        }
      }
    }
  }
  
  return tileCounts, tileCountsSum, tileValuesSum
}

// check a suit's tiles for possible win; helper function for HaveWin
func checkSuitTiles(tmpTileCountsSum int, tmpTileCountsSource []int, suit int) (bool, int, []TileSet) {  
  if VerboseDebug {
    fmt.Printf("[vd] invoke checkSuitTiles %d %v\n", tmpTileCountsSum, tmpTileCountsSource)
  }
  
  tmpTileCounts := make([]int, 10, 10)
  
  tmpSetCount := 0
  tmpSuitSuccess := true
  tmpTileSet := make([]TileSet, 4, 4)
  
  copy(tmpTileCounts, tmpTileCountsSource)
  
  if tmpTileCountsSum == 0 {
    return true, 0, tmpTileSet
  }
  
  if VerboseDebug {
    fmt.Printf("[vd] non-zero sum\n")
  }
  
  if VerboseDebug {
    fmt.Printf("[vd] suit %d\n", suit)
  }
  
  
  
  for j := 1; j <= MaxTileIndex[suit]; j++ {
    if VerboseDebug {
      fmt.Printf("[vd] suit %d index %d\n", suit, j)
    }
    
    if tmpTileCounts[j] == 0 {
      continue
    }
    
    if tmpTileCounts[j] >= 3 {
      tmpTileSet[tmpSetCount] = TileSet{ Kind: "triple", Tiles: UnicodeDisplay[suit][j]+UnicodeDisplay[suit][j]+UnicodeDisplay[suit][j] }
      tmpSetCount++
      tmpTileCounts[j] -= 3
      tmpTileCountsSum -= 3
    }

    if tmpTileCounts[j] < 3 && tmpTileCounts[j] > 0 {
      // remove streets
      for suit != 3 && j <= 7 && tmpTileCounts[j] > 0 && tmpTileCounts[j+1] > 0 && tmpTileCounts[j+2] > 0 {
        tmpTileSet[tmpSetCount] = TileSet{ Kind: "seq", Tiles: UnicodeDisplay[suit][j]+UnicodeDisplay[suit][j+1]+UnicodeDisplay[suit][j+2] }
        tmpSetCount++
        tmpTileCounts[j]--
        tmpTileCounts[j+1]--
        tmpTileCounts[j+2]--
        tmpTileCountsSum -= 3
      }
      if tmpTileCounts[j] != 0 {
        tmpSuitSuccess = false
        return tmpSuitSuccess, tmpSetCount, tmpTileSet
      }
    }
  }
  return tmpSuitSuccess, tmpSetCount, tmpTileSet
}

// determine if the hand has a win, possibly with the presence of an additional tile
func (h PlayerHand) HaveWin(consider Tile, tileSource string) bool {
  if VerboseDebug {
    fmt.Printf("[vd] HaveWin invocation for Player %d with tile %v from %s\n", h.Player, consider, tileSource)
  }
  
  // check for special case
  //if tileSource == "draw" { // tile can be from any source
  // check for special win case
  // suit 4, one of each (1 to 7)
  // suit 1, one and 9
  // suit 2, one and 9
  // suit 3, one and 9
  // one of any previous
  
  specialWin := make([]int, 14, 14)

  if consider != EmptyTile {
    specialWin[SpecialWinTiles[consider.Ud]]++
  }
  
  // suffices to check hidden only as none can be revealed as standalone sets
  for i:= 0; i < len(h.Hidden); i++ {
    if h.Hidden[i] != EmptyTile {
      specialWin[SpecialWinTiles[h.Hidden[i].Ud]]++
    }
  }
  
  haveAtLeastOneOfEach := true
  haveTwoOfOne := false
  
  for i:= 1; i <= 13; i++ {
    if specialWin[i] == 0 {
      haveAtLeastOneOfEach = false
    }
    if specialWin[i] == 2 {
      haveTwoOfOne = true
    }
  }
  
  if haveAtLeastOneOfEach && haveTwoOfOne {
    return true
  }

  //}
  
  // check for ordinary win
  setCount := h.RevealedSets
  eye := ""

  newTileSets := make([]TileSet, 0, 0)
  suitableUse := false // is the tile to be considered included appropriately in the winning set?
  
  tileCounts, tileCountsSum, tileValuesSum := h.CountHiddenTiles(consider)
  
  // either multiple of 3 or multiple of 3 + 2 for each suit; other disqualified
  if VerboseDebug {
    fmt.Printf("[vd] suit sums: %d %d %d %d\n", tileCountsSum[0], tileCountsSum[1], tileCountsSum[2], tileCountsSum[3])
  }
  
  for i:= 0; i < 4; i++ {
    if !(tileCountsSum[i] % 3 == 0 || tileCountsSum[i] % 3 == 2) {
      if VerboseDebug {
        fmt.Printf("[vd] failed suit-level count check\n")
      }
      return false
    }
  }
  
  if VerboseDebug {
    fmt.Printf("[vd] passed suit-level count check\n")
  }
  
  // check each suit separately
  for i:= 0; i < 4; i++ {
    suitSuccess := false
    suitSetCount := 0
    suitTileSets := make([]TileSet, 4,4)
    suitEye := ""
    
    if tileCountsSum[i] == 0 {
      continue
    }
    
    if VerboseDebug {
      fmt.Printf("[vd] starting examination of suit %d\n",i)
    }
    
    if tileCountsSum[i] % 3 == 2 {
      // consider eye removal
      
      if VerboseDebug {
        fmt.Printf("[vd] suit %d might have the set of eyes\n",i)
      }
      
      // if suit 4, remove eye where double exists else fail
      if i == 3 {
        for j:=1; j <= 7; j++ {
          if j == 7 && tileCounts[i][7] != 2 {
            
            if VerboseDebug {
              fmt.Printf("[vd] in suit 3, no doubles were found; missing set of eyes\n")
            }
            return false
          }
          if tileCounts[i][j] == 2 {
            // set tentative eye
            suitEye = UnicodeDisplay[3][j]
            tileCounts[i][j] -= 2
            break
          }
        }
        
        if VerboseDebug {
          fmt.Printf("[vd] check for possible win in suit 3\n")
        }
        
        suitSuccess, suitSetCount, suitTileSets = checkSuitTiles(tileCountsSum[i], tileCounts[i], i)
      } else {
        // uses insight from https://stackoverflow.com/questions/4154960/algorithm-to-find-streets-and-same-kind-in-a-hand/4155177#4155177
        //TODO: simplify duplicated steps
        if (tileValuesSum[i] % 3 == 0 && tileCounts[i][3] >= 2) || (tileValuesSum[i] % 3 == 1 && tileCounts[i][2] >= 2) || (tileValuesSum[i] % 3 == 2 && tileCounts[i][1] >= 2) {
          tmpTileCountsSum := tileCountsSum[i] - 2
          tmpTileCounts := make([]int,10,10)
          copy(tmpTileCounts, tileCounts[i])
          
          if VerboseDebug {
            fmt.Printf("[vd] suit %d: consider the first set of eyes\n",i)
          }
          
          // set tentative eye
          switch tileValuesSum[i] % 3 {
            case 0: {
              suitEye = UnicodeDisplay[i][3]
              tmpTileCounts[3] -= 2
            }
            case 1: {
              suitEye = UnicodeDisplay[i][2]
              tmpTileCounts[2] -= 2
            }
            case 2: {
              suitEye = UnicodeDisplay[i][1]
              tmpTileCounts[1] -= 2
            }
          }
          
          suitSuccess, suitSetCount, suitTileSets = checkSuitTiles(tmpTileCountsSum, tmpTileCounts, i)
        }
        
        if !suitSuccess && i != 3 {
          if (tileValuesSum[i] % 3 == 0 && tileCounts[i][6] >= 2) || (tileValuesSum[i] % 3 == 1 && tileCounts[i][5] >= 2) || (tileValuesSum[i] % 3 == 2 && tileCounts[i][4] >= 2) {
            tmpTileCountsSum := tileCountsSum[i] - 2
            tmpTileCounts := make([]int,10,10)
            copy(tmpTileCounts, tileCounts[i])
            
            // set tentative eye
            switch tileValuesSum[i] % 3 {
              case 0: {
                suitEye = UnicodeDisplay[i][6]
                tmpTileCounts[6] -= 2
              }
              case 1: {
                suitEye = UnicodeDisplay[i][5]
                tmpTileCounts[5] -= 2
              }
              case 2: {
                suitEye = UnicodeDisplay[i][4]
                tmpTileCounts[4] -= 2
              }
            }
            
            if VerboseDebug {
              fmt.Printf("[vd] suit %d: consider the second set of eyes\n",i)
            }

            suitSuccess, suitSetCount, suitTileSets = checkSuitTiles(tmpTileCountsSum, tmpTileCounts, i)
          }
        }
        
        if !suitSuccess && i != 3 {
          if (tileValuesSum[i] % 3 == 0 && tileCounts[i][9] >= 2) || (tileValuesSum[i] % 3 == 1 && tileCounts[i][8] >= 2) || (tileValuesSum[i] % 3 == 2 && tileCounts[i][7] >= 2) {
            tmpTileCountsSum := tileCountsSum[i] - 2
            tmpTileCounts := make([]int,10,10)
            copy(tmpTileCounts, tileCounts[i])
            
            // set tentative eye
            switch tileValuesSum[i] % 3 {
              case 0: {
                suitEye = UnicodeDisplay[i][9]
                tmpTileCounts[9] -= 2
              }
              case 1: {
                suitEye = UnicodeDisplay[i][8]
                tmpTileCounts[8] -= 2
              }
              case 2: {
                suitEye = UnicodeDisplay[i][7]
                tmpTileCounts[7] -= 2
              }
            }
            
            if VerboseDebug {
              fmt.Printf("[vd] suit %d: consider the third set of eyes\n",i)
            }

            suitSuccess, suitSetCount, suitTileSets = checkSuitTiles(tmpTileCountsSum, tmpTileCounts, i)
          }
        }

      }
    } else {
      if VerboseDebug {
        fmt.Printf("[vd] suit %d has no eyes; standard check\n",i)
      }
        
      suitSuccess, suitSetCount, suitTileSets = checkSuitTiles(tileCountsSum[i], tileCounts[i], i)
    }
    
    if suitSuccess == false {
      // one suit fails, cannot be a win
      if VerboseDebug {
        fmt.Printf("[vd] Suit %d failed the win test; cannot have a win\n", i)
      }
      return false
    } else {
      
      if VerboseDebug {
        fmt.Printf("[vd] suit %d passed with %d sets and eye of %v\n", i, suitSetCount, suitEye)
        fmt.Printf("[vd] tile sets: %v\n", suitTileSets)
      }
      
      setCount += suitSetCount
      // append tile sets
      for j:= 0; j < suitSetCount; j++ {
        newTileSets = append(newTileSets, suitTileSets[j])
      }
      // update eye if populated
      if len(suitEye) > 0 {
        eye = suitEye
      }
      
    }
  }
  
  if VerboseDebug {
    fmt.Printf("[vd] current set count: %d; pending suitability check\n", setCount)
  }
  
  if tileSource == "previous" || tileSource == "other" {
    // check for win that involves tile as pong or sequence in hidden part if previous
    // check for win that involes the tile as pong if other

    for i:= 0; i < setCount-h.RevealedSets; i++ {
      if newTileSets[i].Kind == "triple" && strings.Contains(newTileSets[i].Tiles, consider.Ud) {
        suitableUse = true
        break
      } else if tileSource == "previous" && newTileSets[i].Kind == "seq" && strings.Contains(newTileSets[i].Tiles, consider.Ud) {
        suitableUse = true
      }
    }
        
    if !suitableUse && consider.Ud == eye {
      suitableUse = true
    }
    
    if VerboseDebug {
      fmt.Printf("[vd] Suitable use outcome: %v, %v, %v\n", suitableUse, consider.Ud, eye)
    }

  } else {
    suitableUse = true
  }
  
  if setCount == 4 && suitableUse == true {
    if VerboseDebug {
      fmt.Printf("[vd] Possible WIN: %v: %v\n", eye, newTileSets)
    }
    return true
  }

  return false
}

// check to see if the player has a set of four, perhaps, with an optional extra tile
func (h PlayerHand) HaveKong(consider Tile, tileSource string) (bool, []TileSet) {
  kongFound := false
  kongSets := make([]TileSet, 0, 0)
  
  tileCounts, _, _ := h.CountHiddenTiles(consider)
  
  for i:= 0; i < 4; i++ {
    for j := 1; j < 10; j++ {
      if tileCounts[i][j] == 4 && (consider == EmptyTile || (consider.Suit-1 == i && consider.Value == j)) {
        kongFound = true
        kongSets = append(kongSets, TileSet{ Kind: "kong", Tiles: UnicodeDisplay[i][j]+UnicodeDisplay[i][j]+UnicodeDisplay[i][j]+UnicodeDisplay[i][j] })
      }
      if i == 3 && j == 7 {
        break
      }
    }
  }
  
  for i:= 0; i < h.RevealedSets; i++ {
    if h.RevealedTileSets[i].Kind == "triple" && tileSource == "draw" {
      for j:= 0; j < 4; j++ {
        for k := 1; k < 10; k++ {
          if tileCounts[j][k] == 1 && strings.Contains(h.RevealedTileSets[i].Tiles, UnicodeDisplay[j][k]) {
            kongFound = true
            kongSets = append(kongSets, TileSet{ Kind: "kong", Tiles: UnicodeDisplay[j][k]+UnicodeDisplay[j][k]+UnicodeDisplay[j][k]+UnicodeDisplay[j][k] })
          }
          if j == 3 && k == 7 {
            break
          }
        }
      }
      
    }
  }
  
  return kongFound, kongSets
}

// check to see if the player has a set of three with an extra tile
func (h PlayerHand) HavePong(consider Tile, tileSource string) (bool, string) {
  if tileSource != "previous" && tileSource != "other" {
    return false, ""
  }
  
  tileCounts, _, _ := h.CountHiddenTiles(consider)
  
  for i:= 0; i < 4; i++ {
    for j := 1; j < 10; j++ {
      if tileCounts[i][j] == 3 && consider != EmptyTile && consider.Suit -1 == i && consider.Value == j {
        return true, UnicodeDisplay[i][j]
      }
      if i == 3 && j == 7 {
        break
      }
    }
  }
  
  return false,""
}

// check if the next player can form a sequence with the discard
func (h PlayerHand) HaveSeq(consider Tile, tileSource string) (bool, []TileSet) {
  seqFound := false
  seqSets := make([]TileSet, 0, 0)
  
  if tileSource != "previous" {
    return seqFound, seqSets
  }
  
  tileCounts, _, _ := h.CountHiddenTiles(consider)
  
  for i:= 0; i < 3; i++ {
    for j := 1; j < 7; j++ {
      if tileCounts[i][j] > 0 && tileCounts[i][j+1] > 0 && tileCounts[i][j+2] > 0 && consider != EmptyTile && consider.Suit -1 == i && (consider.Value == j || consider.Value == j+1 || consider.Value == j+2) {
        seqFound = true
        seqSets = append(seqSets, TileSet{ Kind: "seq", Tiles: UnicodeDisplay[i][j]+UnicodeDisplay[i][j+1]+UnicodeDisplay[i][j+2] })
      }
    }
  }
  
  return seqFound, seqSets
}

// computer player: take the win?
// naively, yes
func (h PlayerHand) TakeWin(discard []DiscardedTile, considerLastDiscard bool, hands []PlayerHand) string {
  return "y"
}        

// computer player: take kong?
// naively, yes
func (h PlayerHand) TakeKong(discard []DiscardedTile, considerLastDiscard bool, hands []PlayerHand) string {
  return "y"
}

// computer player: take pong?
// naively, yes
func (h PlayerHand) TakePong(discard []DiscardedTile, considerLastDiscard bool, hands []PlayerHand) string {
  return "y"
}

// computer player: take seq?
// naively, take the first one
func (h PlayerHand) TakeSeq(discard []DiscardedTile, considerLastDiscard bool, hands []PlayerHand, options []TileSet) string {
  return "0"
}      

// computer player: what to discard?
// naively, the first tile
func (h PlayerHand) Discard(discard DiscardPile, considerLastDiscard bool, hands []PlayerHand) string {
  // count internal hand
  tileCounts, tileCountsSum, tileValuesSum := h.CountHiddenTiles(EmptyTile)
  // count discarded
  dtileCounts, dtileCountsSum, dtileValuesSum := discard.CountDiscardTiles(false)
  // count unusable tiles (aside from discard)
  u0tileCounts, u0tileCountsSum, u0tileValuesSum := hands[0].CountPublicSetTiles()
  u1tileCounts, u1tileCountsSum, u1tileValuesSum := hands[1].CountPublicSetTiles()
  u2tileCounts, u2tileCountsSum, u2tileValuesSum := hands[2].CountPublicSetTiles()
  u3tileCounts, u3tileCountsSum, u3tileValuesSum := hands[3].CountPublicSetTiles()

  // merge unavailable tiles
  unavailableTileCounts := make([][]int, 4, 4)
  for i := 0; i < 4; i++ {
    unavailableTileCounts[i] = make([]int, 10, 10)
  } // allocating an extra two positions for the honor suit

  for i := 0; i < 4; i++ {
    for j := 0; j < 10; j++ {
      unavailableTileCounts[i][j] += dtileCounts[i][j]
      unavailableTileCounts[i][j] += u0tileCounts[i][j]
      unavailableTileCounts[i][j] += u1tileCounts[i][j]
      unavailableTileCounts[i][j] += u2tileCounts[i][j]
      unavailableTileCounts[i][j] += u3tileCounts[i][j]
    }
  }

  if (VerboseDebug) {
    fmt.Println("[vd] internal hand", tileCounts, tileCountsSum, tileValuesSum)
    fmt.Println("[vd] discard", dtileCounts, dtileCountsSum, dtileValuesSum)
    fmt.Println("[vd] p0 public sets", u0tileCounts, u0tileCountsSum, u0tileValuesSum)
    fmt.Println("[vd] p1 public sets", u1tileCounts, u1tileCountsSum, u1tileValuesSum)
    fmt.Println("[vd] p2 public sets", u2tileCounts, u2tileCountsSum, u2tileValuesSum)
    fmt.Println("[vd] p3 public sets", u3tileCounts, u3tileCountsSum, u3tileValuesSum)
    fmt.Println("[vd] unavailable tiles", unavailableTileCounts)    
  }
  
  var hiddenTileCount int
  for _, tile := range h.Hidden {
    if tile != EmptyTile {
      hiddenTileCount++
    }
  }
  
  // remove intact sets from consideration
  // note: misses options (e.g., if 1-2-3-4 is present, 1-2-3 will be preferentially removed)
  for i := 0; i < 4; i++ {
    for j := 1; j < 10; j++ {
      if tileCounts[i][j] >= 3 {
        tileCounts[i][j] -= 3
        
        unavailableTileCounts[i][j] += 3
        
        hiddenTileCount -= 3
      } else if i != 3 && j <= 7 && tileCounts[i][j] > 0 && tileCounts[i][j+1] > 0 && tileCounts[i][j+2] > 0 {
        tileCounts[i][j]--
        tileCounts[i][j+1]--
        tileCounts[i][j+2]--
        
        unavailableTileCounts[i][j]++
        unavailableTileCounts[i][j+1]++
        unavailableTileCounts[i][j+2]++
        
        hiddenTileCount -= 3
      }
    }
  }
  
  // pairs (unless blocked) > sequential twos in middle > sequential twos at end > gapped twos > singles
  
  // cannot have a pair remaining; otherwise, it would be a win
  for i := 0; i < 4; i++ {
    for j := 1; j < 10; j++ {
      if hiddenTileCount > 2 && tileCounts[i][j] >= 2 && unavailableTileCounts[i][j] < 2 {
        tileCounts[i][j] -= 2
        
        unavailableTileCounts[i][j] += 2
        
        hiddenTileCount -= 2
      }
    }
  }
  
  // check sequential twos in middle
  for i := 0; i < 3; i++ {
    for j := 2; j < 7; j++ {
      if hiddenTileCount > 2 && tileCounts[i][j] >= 1 && tileCounts[i][j+1] >= 1 && (unavailableTileCounts[i][j+2] < 3 || unavailableTileCounts[i][j-1] < 3) {
        tileCounts[i][j]--
        tileCounts[i][j+1]--
        
        unavailableTileCounts[i][j]++
        unavailableTileCounts[i][j+1]++
                  
        hiddenTileCount -= 2
      }
    }
  }
  
  // check twos at edges (start)
  for i := 0; i < 3; i++ {
    if hiddenTileCount > 2 && tileCounts[i][1] >= 1 && tileCounts[i][2] >= 1 && unavailableTileCounts[i][3] < 3 {
      tileCounts[i][1]--
      tileCounts[i][2]--
      
      unavailableTileCounts[i][1]++
      unavailableTileCounts[i][2]++
                
      hiddenTileCount -= 2
    }
  }
  
  // check twos at edges (end)
  for i := 0; i < 3; i++ {
    if hiddenTileCount > 2 && tileCounts[i][8] >= 1 && tileCounts[i][9] >= 1 && unavailableTileCounts[i][7] < 3 {
      tileCounts[i][8]--
      tileCounts[i][9]--
      
      unavailableTileCounts[i][8]++
      unavailableTileCounts[i][9]++
                
      hiddenTileCount -= 2
    }
  }
  
  // check for gapped sequence
  for i := 0; i < 3; i++ {
    for j := 1; j < 8; j++ {
      if hiddenTileCount > 2 && tileCounts[i][j] >= 1 && tileCounts[i][j+2] >= 1 && unavailableTileCounts[i][j+1] < 3 {
        tileCounts[i][j]--
        tileCounts[i][j+2]--
        
        unavailableTileCounts[i][j]++
        unavailableTileCounts[i][j+2]++
                  
        hiddenTileCount -= 2
      }
    }
  }
  
  // choose randomly for now
  options := make([]string,0,0)
  optionCounter := 0
  
  for i := 0; i < 4; i++ {
    for j := 1; j < 10; j++ {
      for k := 0; k < tileCounts[i][j]; k++ {
        options = append(options,UnicodeDisplay[i][j])
        optionCounter++
      }
    }
  }
  
  if VerboseDebug {
    fmt.Println("[vd] suggested discard options: ",options, optionCounter, hiddenTileCount)
  }
  // TODO: centralize rand
  insecureRand.Seed(12345);
  choice := insecureRand.Intn(optionCounter)
  //fmt.Println(tileCounts, tileCountsSum, tileValuesSum)
  handPosition := 0
  
  h.Sort()
  
  for i := 0; i < len(h.Hidden); i++ {
    if h.Hidden[i].Ud == options[choice] {
      if VerboseDebug {
        fmt.Println("[vd] suggested discard: ", options[choice])
      }
      handPosition = i
    }
  }
  
  return strconv.Itoa(handPosition)
}


// process the initial deal; as presentation is important, the dealing processing is followed strictly
func (g *Game) InitialDeal() {
  // rounds of dealing
  for k:= 0; k < 3; k++ {
    // each player
    for i:= 0; i < PlayersInGame; i++ {
      // each of four tiles
      for j:= 0; j < 4; j++ {
        curTile, err := g.GetInitialTile(k, i, j)
        if err != nil {
          log.Fatal(err)
        }
        err = g.Hands[(i+g.CurrentPlayer)%4].Receive(curTile)
        if err != nil {
          log.Fatal(err)
        }
      }
    }
  }
  
  // special allocation for dealer
  curTile, err := g.GetInitialTile(3, 0, 0)
  if err != nil {
    log.Fatal(err)
  }
  
  err = g.Hands[g.CurrentPlayer].Receive(curTile)
  if err != nil {
    log.Fatal(err)
  }

  curTile, err = g.GetInitialTile(3, 1, 0)
  if err != nil {
    log.Fatal(err)
  }
  
  err = g.Hands[g.CurrentPlayer].Receive(curTile)
  if err != nil {
    log.Fatal(err)
  }
  
  // final allocation for others
  for i := 1; i < PlayersInGame; i++ {
    curTile, err := g.GetInitialTile(3, 0, i)
    if err != nil {
      log.Fatal(err)
    }

    err = g.Hands[(i+g.CurrentPlayer)%4].Receive(curTile)
    if err != nil {
      log.Fatal(err)
    }
  }
}

// process special tiles occurring as part of the initial deal; this completes the initialization and is the begin object
func (g *Game) InitialHandleSpecialTiles() {
  for i := 0; i < PlayersInGame; i++ {
    for g.Hands[(i+g.CurrentPlayer)%4].HasUnreplacedSpecialTile() {
      specialTile, err := g.Hands[(i+g.CurrentPlayer)%4].GetFirstSpecialTile()
      if err != nil {
        log.Fatal(err)
      }
      
      err = g.Hands[(i+g.CurrentPlayer)%4].RevealSpecialTile(specialTile)
      if err != nil {
        log.Fatal(err)
      }
      
      curTile, err := g.GetNewTile(&g.ReplacementPointer, true)
      if err != nil {
        log.Fatal(err)
      }
            
      g.Hands[(i+g.CurrentPlayer)%4].Receive(curTile)
      if err != nil {
        log.Fatal(err)
      }
      
      if VerboseDebug {
        fmt.Printf("[vd] Player %d had special tile %v, which was replaced with tile %v\n", g.Hands[(i+g.CurrentPlayer)%4].Player, specialTile, curTile)
      }

    }
  }
}
