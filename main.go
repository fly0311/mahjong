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

package main

import (
  mahjong "./mahjong"
  "flag"
  "log"
  "io"
  "io/ioutil"
  "os"
  //"fmt"
)

func main() {
  p := -1
  game := 0
  
  singlePlayerMode := flag.Bool("singlePlayer", false, "single player mode with computer players? [bool]")
  logFile := flag.String("logFile", "", "log file for game [file path]")
    
  flag.Parse()
  
  var computerPlayers []bool = make([]bool, 4, 4)
  if *singlePlayerMode {
    computerPlayers[1] = *singlePlayerMode
    computerPlayers[2] = *singlePlayerMode
    computerPlayers[3] = *singlePlayerMode
  }

  // logging boilerplate: https://www.goinggo.net/2013/11/using-log-package-in-go.html
  var outputLogDestination io.Writer
  var err error
  if *logFile == "" {
    outputLogDestination = ioutil.Discard
  } else {
    outputLogDestination, err = os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
      log.Fatalln("Could not open for writing: ", *logFile, ":", err)
    }
  }
  
  logInstance := log.New(outputLogDestination, "ACTION: ", 0)
    
  // single game mode for now
  for game < 1 {
    // new game
    currentGame := mahjong.New()
    currentGame.OutputLog = logInstance
    currentGame.Initialize(p, computerPlayers)

    // return outcome
    _, p = currentGame.BeginGame()

    game++
  }
}
