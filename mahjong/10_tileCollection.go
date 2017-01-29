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

// handler for tile operations
package mahjong

import(
  "fmt"
  // golang: handle namespace collision
  insecureRand "math/rand"
  "crypto/rand"
  "math/big"
  "errors"
)

// nomenclature loosely follows https://en.wikipedia.org/wiki/Mahjong#Old_Hong_Kong_Mahjong

// # Tile
// tile; uninitialized tile is considered empty
type Tile struct {
  // golang: variables have default values; use 0,0 tile as default, resulting in 1-based indexing to handle default uninitialized case

  // 1-3: dots, bamboo, characters; 4: honors; 5 bonus (retitled special)
  Suit int
  
  // for standard suits, 1-9; 
  // for honors: east, south, west, north, red, green, white; 
  // for bonus: flowers 1-4; seasons spring to fall
  Value int
  
  // for tile tracking
  Id int
  
  // unicode display
  Ud string
}

// uninitialized tile
var EmptyTile Tile

// golang: native support for unicode!
// return tile serial and human readable character
func (t Tile) String() string {
  return fmt.Sprintf("[%03d%v]", t.Id, t.Ud)
}

// return only human readable character
func (t Tile) UdString() string {
  return fmt.Sprintf("%v", t.Ud)
}

// return if a tile is a special tile
func (t Tile) IsSpecial() bool {
  if t.Suit == 5 {
    return true
  }
  return false
}

// # tileCollection
type TileCollection []Tile
// translation to unicode characters
var UnicodeDisplay [][]string
// show verbose debug messages
var VerboseDebug bool
// tiles needed for a special win
var SpecialWinTiles map[string]int
// is rand deterministic?
var DeterministicRand bool

const (
  TilesInGame = 144
)

// discarded tile
type DiscardedTile struct {
  Player int
  Item Tile
}

// discards
type DiscardPile []DiscardedTile
  
// populate tileCollection
func init() {
  VerboseDebug = false
  DeterministicRand = false
  
  // allocate tiles needed for a special win
  SpecialWinTiles = make(map[string]int)
  // zero to handle no match case
  SpecialWinTiles["🀀"] = 1
  SpecialWinTiles["🀁"] = 2
  SpecialWinTiles["🀂"] = 3
  SpecialWinTiles["🀃"] = 4
  SpecialWinTiles["🀄"] = 5
  SpecialWinTiles["🀅"] = 6
  SpecialWinTiles["🀆"] = 7
  SpecialWinTiles["🀙"] = 8
  SpecialWinTiles["🀡"] = 9
  SpecialWinTiles["🀐"] = 10
  SpecialWinTiles["🀘"] = 11
  SpecialWinTiles["🀇"] = 12
  SpecialWinTiles["🀏"] = 13
  
  // seed deterministic generator
  insecureRand.Seed(12345);
  
  // allocate Unicode display content
  UnicodeDisplay = make([][]string, 5, 5)
  UnicodeDisplay[0] = []string {"", "🀙", "🀚", "🀛", "🀜", "🀝", "🀞", "🀟", "🀠", "🀡"}
  UnicodeDisplay[1] = []string {"", "🀐", "🀑", "🀒", "🀓", "🀔", "🀕", "🀖", "🀗", "🀘"}
  UnicodeDisplay[2] = []string {"", "🀇", "🀈", "🀉", "🀊", "🀋", "🀌", "🀍", "🀎", "🀏"}
  UnicodeDisplay[3] = []string {"", "🀀", "🀁", "🀂", "🀃", "🀄", "🀅", "🀆"}
  UnicodeDisplay[4] = []string {"", "🀢", "🀣", "🀤", "🀥", "🀦", "🀧", "🀨", "🀩"}
}

// shuffle undealt tiles (only if not previously shuffled)
func (g *Game) Shuffle(deterministic bool) error {
  if !g.Shuffled {
    g.Shuffled = true
    for i := range g.Undealt {
      var k int
      if deterministic {
        k = insecureRand.Intn(i+1)
      } else {
        j, err := rand.Int(rand.Reader, big.NewInt(int64(i)+1)) // note 0, excluding max
        if err != nil {
          return err
        }
        k = int(j.Int64())
      }
      if i != k {
        g.Undealt[i], g.Undealt[k] = g.Undealt[k], g.Undealt[i]
      }
    }
  } else {
    return errors.New("Shuffling already attempted")
  }
  return nil
}

// output undealt tiles and positions
func (g *Game) OutputUndealtTiles() {
  for i := 0; i < TilesInGame; i++ {
    if g.Undealt[i] != EmptyTile {
      fmt.Printf("%3d: %v\n", i, g.Undealt[i])
    } else {
      fmt.Printf("%3d: empty\n", i)
    }
  }
}

func (g *Game) SetDealLocations(diceRoll int) error {
  if diceRoll < 3 || diceRoll > 18 {
    return fmt.Errorf("Sum of three dice is %d, not between 3 and 18", diceRoll)
  }
  if g.DrawLocationsSet {
    return errors.New("Deal locations already set")
  } else {
    g.DrawLocationsSet = true
  }
  
  // only valid if in the east location
  g.AllocationStart = ((diceRoll-1)%4)*36+(diceRoll)*2+g.StartPlayer*36
  
  g.DrawPointer = (g.AllocationStart + 16*3 + 5) % 144
  g.ReplacementPointer = g.AllocationStart-1 % 144
  
  return nil
}

func (g *Game) GetInitialTile(round int, player int, current int) (Tile, error) {
    // have tiles to deal?
  if g.UndealtTileCount < 1 {
    return EmptyTile, fmt.Errorf("no tiles to deal; undealt tile count at %d", g.UndealtTileCount)
  }
  
  allocationPosition := (g.AllocationStart + 16*round + 4*player + current) % 144
  if g.Undealt[allocationPosition] == EmptyTile {
    return EmptyTile, fmt.Errorf("round %d, player %d, item %d with an AllocationStart of %d yields %d, which is empty", round, player, current, g.AllocationStart, allocationPosition)
  }
  
  selection := EmptyTile
  selection, g.Undealt[allocationPosition] = g.Undealt[allocationPosition], EmptyTile
  
  g.UndealtTileCount--
    
  return selection, nil
}

// retrieve new tile for both draw and replacement
func (g *Game) GetNewTile(pointer* int, replacement bool) (Tile, error) {
  if g.UndealtTileCount < 1 {
    return EmptyTile, errors.New("no more tiles to deal")
  }
  
  if (*pointer) < 0 {
    return EmptyTile, errors.New("uninitialized pointer?")
  }
  
  (*pointer) = (*pointer) % 144

  if (*pointer) < 0 || g.Undealt[(*pointer)] == EmptyTile {
    return EmptyTile, fmt.Errorf("expected a tile at position %d, which is empty", (*pointer))
  }
  
  var newTile Tile
  newTile, g.Undealt[(*pointer)] = g.Undealt[(*pointer)], EmptyTile
  g.UndealtTileCount--
  
  if replacement {
    (*pointer)--
  } else {
    (*pointer)++
  }
  
  return newTile, nil
}

// # dice function
// roll one dice
func RollOneDice(deterministic bool) int {
  if deterministic {
    return insecureRand.Intn(6)+1
  }else {
    i, _ := rand.Int(rand.Reader, big.NewInt(int64(6)))
    return int(i.Int64())+1
  }
}

// # discard pile
func (g *Game) OutputDiscardedTiles() {
  if len(g.Discard) == 0 {
    fmt.Printf("D: [no tiles in discard]\n")
  } else {
    for i := 0; i <= len(g.Discard)/8; i++ {
      for j := 0; j < 8; j++ {
        k := i*8+j
        if k >= len(g.Discard) {
          fmt.Println()
          break
        }
        if j == 0 {
          fmt.Printf("D: ")
        }
        fmt.Printf("(%v-%d)", g.Discard[k].Item.Ud, g.Discard[k].Player)
        if j == 7 {
          fmt.Println()
        }
      }
    }
    fmt.Printf("\n")
  }
}
