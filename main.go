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

package main

import (
  mahjong "./mahjong"
)

func main() {
  p := -1
  game := 0
  
  // single game mode for now
  for game < 1 {
    // new game
    currentGame := mahjong.New()
    currentGame.Initialize(p)

    // return outcome
    _, p = currentGame.BeginGame()

    game++
  }
}
