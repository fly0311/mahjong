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
  "testing"
)

func TestTileCollectionPopulation(t *testing.T) {
  tileCounter := make(map[int]int)
  tileCount := 0

  for i := 0; i < len(gt.Undealt); i++ {
    tileCounter[gt.Undealt[i].Id]++
    tileCount++
  }
  
  if tileCount != 144 {
    t.Errorf("Initialization failure; fewer than 144 tiles were counted.")
  }
  
  // note checking serials
  for i := 1; i <= 144; i++ {
    if tileCounter[i] != 1 {
      t.Errorf("Initialization failure; serial %d has %d instances (only one should exist).", i, tileCounter[i])
    }
  }
}

// TODO: ensure shuffle fails when already attempted
