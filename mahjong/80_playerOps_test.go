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

type TestHand struct {
  Tiles string
  Relationship string
  Outcome bool
}

func TestWinningHands(t *testing.T) {
  var testCases []TestHand
  
  // special win
  testCases = append(testCases, TestHand{ Tiles:"🀀🀁🀂🀃🀄🀅🀆🀙🀐🀇🀡🀘🀏🀀;", Relationship: "draw", Outcome: true })
  testCases = append(testCases, TestHand{ Tiles:"🀀🀁🀂🀃🀄🀅🀆🀙🀐🀇🀡🀘🀏;🀀", Relationship: "previous", Outcome: true })
  testCases = append(testCases, TestHand{ Tiles:"🀀🀁🀂🀃🀄🀅🀆🀙🀐🀇🀡🀘🀏;🀀", Relationship: "other", Outcome: true })
  
  // generic
  testCases = append(testCases, TestHand{ Tiles:"🀑🀒🀓🀉🀉🀉🀝🀞🀒🀒🀟🀆🀆🀆;", Relationship: "draw", Outcome: true })
  // eye
  testCases = append(testCases, TestHand{ Tiles:"🀑🀒🀓🀉🀉🀉🀝🀞🀒🀟🀆🀆🀆;🀒", Relationship: "other", Outcome: true })
  // set
  testCases = append(testCases, TestHand{ Tiles:"🀑🀒🀓🀉🀉🀝🀞🀒🀒🀟🀆🀆🀆;🀉", Relationship: "other", Outcome: true })
  // seq
  testCases = append(testCases, TestHand{ Tiles:"🀑🀒🀓🀉🀉🀉🀝🀒🀒🀟🀆🀆🀆;🀞", Relationship: "previous", Outcome: true })
  
  // not a win
  testCases = append(testCases, TestHand{ Tiles:"🀑🀒🀓🀉🀉🀉🀝🀒🀒🀟🀆🀆🀆;🀞", Relationship: "other", Outcome: false })
  testCases = append(testCases, TestHand{ Tiles:"🀑🀒🀓🀉🀉🀇🀝🀞🀒🀒🀟🀆🀆🀆;", Relationship: "draw", Outcome: false })
  
  for i := 0; i < len(testCases); i++ {
    testHand, testTile := gt.TestHandMaker(testCases[i].Tiles)
    if testCases[i].Outcome != testHand.HaveWin(testTile, testCases[i].Relationship) {
      t.Errorf("%v with additional tile %v arising from %s should have been a %v, but was not", testHand, testTile, testCases[i].Relationship, testCases[i].Outcome)
    }
  }
}

func TestSequenceCheck(t *testing.T) {
  var testCases []TestHand
  
  // invalid invocation; only applicable to a previous relationship
  testCases = append(testCases, TestHand{ Tiles:"🀑🀒🀓🀉🀉🀇🀝🀞🀒🀒🀟🀆🀆🀆;", Relationship: "draw", Outcome: false })
  testCases = append(testCases, TestHand{ Tiles:"🀑🀓🀉🀉🀇🀝🀞🀒🀒🀟🀆🀆🀆;🀒", Relationship: "other", Outcome: false })
  
  // previous
  testCases = append(testCases, TestHand{ Tiles:"🀑🀓🀉🀉🀇🀝🀞🀒🀒🀟🀆🀆🀆;🀒", Relationship: "previous", Outcome: true })
  testCases = append(testCases, TestHand{ Tiles:"🀇🀈🀉🀊🀋🀌🀍🀎🀏🀏🀆🀆🀆;🀊", Relationship: "previous", Outcome: true })
  
  for i := 0; i < len(testCases); i++ {
    testHand, testTile := gt.TestHandMaker(testCases[i].Tiles)
    if outcome, _ := testHand.HaveSeq(testTile, testCases[i].Relationship); outcome != testCases[i].Outcome {
      t.Errorf("%v with additional tile %v arising from %s should have been a %v, but was not", testHand, testTile, testCases[i].Relationship, testCases[i].Outcome)
    }
  }
  
  // manually check the multiple options
  testHand, testTile := gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀍🀎🀏🀏🀆🀆🀆;🀊")
  outcome, setOptions := testHand.HaveSeq(testTile, "previous")
  if outcome {
    expectedSets := make(map[string]int)
    counter := 0
    for i := 0; i < len(setOptions); i++ {
      expectedSets[setOptions[i].Tiles]++
      counter++
    }
    if counter != 3 || expectedSets["🀈🀉🀊"] != 1 || expectedSets["🀉🀊🀋"] != 1 || expectedSets["🀊🀋🀌"] != 1 {
      t.Errorf("not enough options returned or too many")
    }
  } else {
    t.Errorf("sequence missed in the manual multi option check")
  }
}

func TestKongCheck(t *testing.T) {  
  // one kong in draw case
  testHand, testTile := gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀍🀎🀏🀏🀆🀆🀆🀆;")
  outcome, setOptions := testHand.HaveKong(testTile, "draw")
  if !outcome {
    t.Errorf("a kong was missed: one kong in draw case")
  }
  if !(len(setOptions) == 1 && setOptions[0].Tiles == "🀆🀆🀆🀆") {
    t.Errorf("kong count was missed: one kong in draw case")
  }
  
  // two kong in draw case
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀏🀏🀏🀏🀆🀆🀆🀆;")
  outcome, setOptions = testHand.HaveKong(testTile, "draw")
  if outcome {
    expectedSets := make(map[string]int)
    counter := 0
    for i := 0; i < len(setOptions); i++ {
      expectedSets[setOptions[i].Tiles]++
      counter++
    }
    if counter != 2 || expectedSets["🀆🀆🀆🀆"] != 1 || expectedSets["🀏🀏🀏🀏"] != 1 {
      t.Errorf("not enough options returned or too many for two kong in draw case")
    }
  } else {
    t.Errorf("kong was missed: two kong in draw case")
  }
  
  // one pre-existing kong and one new kong in other case
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀏🀏🀏🀏🀆🀆🀆;🀆")
  outcome, setOptions = testHand.HaveKong(testTile, "previous")
  if !outcome {
    t.Errorf("a kong was missed: one pre-existing kong and one new kong in previous case")
  }
  if !(len(setOptions) == 1 && setOptions[0].Tiles == "🀆🀆🀆🀆") {
    t.Errorf("kong count was missed: one pre-existing kong and one new kong in previous case")
  }

  // one pre-existing kong and one new kong in other case
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀏🀏🀏🀏🀆🀆🀆;🀆")
  outcome, setOptions = testHand.HaveKong(testTile, "other")
  if !outcome {
    t.Errorf("a kong was missed: one pre-existing kong and one new kong in other case")
  }
  if !(len(setOptions) == 1 && setOptions[0].Tiles == "🀆🀆🀆🀆") {
    t.Errorf("kong count was missed: one pre-existing kong and one new kong in other case")
  }
  
  // one new kong in previous case
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀌🀏🀏🀏🀆🀆🀆;🀆")
  outcome, setOptions = testHand.HaveKong(testTile, "previous")
  if !outcome {
    t.Errorf("a kong was missed: one new kong in previous case")
  }
  if !(len(setOptions) == 1 && setOptions[0].Tiles == "🀆🀆🀆🀆") {
    t.Errorf("kong count was missed: one new kong in previous case")
  }

  // one new kong in other case
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀌🀏🀏🀏🀆🀆🀆;🀆")
  outcome, setOptions = testHand.HaveKong(testTile, "other")
  if !outcome {
    t.Errorf("a kong was missed: one new kong in other case")
  }
  if !(len(setOptions) == 1 && setOptions[0].Tiles == "🀆🀆🀆🀆") {
    t.Errorf("kong count was missed: one new kong in other case")
  }

  // no new kong in other case
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀌🀏🀏🀆🀆🀆🀆;🀏")
  outcome, setOptions = testHand.HaveKong(testTile, "previous")
  if outcome {
    t.Errorf("a kong was misidentified: no new kong in previous case")
  }
  
  // no new kong in other case
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀌🀏🀏🀆🀆🀆🀆;🀏")
  outcome, setOptions = testHand.HaveKong(testTile, "other")
  if outcome {
    t.Errorf("a kong was misidentified: no new kong in other case")
  }
  
  // permitted addition to a revealed pong via draw
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀌🀏🀏🀏🀆;")
  testHand.RevealedSets = 1
  testHand.RevealedTileSets = append(testHand.RevealedTileSets, TileSet{ Kind: "triple", Tiles: "🀆🀆🀆" })
  outcome, setOptions = testHand.HaveKong(testTile, "draw")
  if !outcome {
    t.Errorf("permitted addition to a revealed pong via draw missed")
  }
  
  // unpermitted addition to a revealed pong via previous
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀌🀏🀏🀏;🀆")
  testHand.RevealedSets = 1
  testHand.RevealedTileSets = append(testHand.RevealedTileSets, TileSet{ Kind: "triple", Tiles: "🀆🀆🀆" })
  outcome, setOptions = testHand.HaveKong(testTile, "previous")
  if outcome {
    t.Errorf("unpermitted addition to a revealed pong via previous misidentified")
  }
  
  // unpermitted addition to a revealed pong via other
  testHand, testTile = gt.TestHandMaker("🀇🀈🀉🀊🀋🀌🀌🀏🀏🀏;🀆")
  testHand.RevealedSets = 1
  testHand.RevealedTileSets = append(testHand.RevealedTileSets, TileSet{ Kind: "triple", Tiles: "🀆🀆🀆" })
  outcome, setOptions = testHand.HaveKong(testTile, "other")
  if outcome {
    t.Errorf("unpermitted addition to a revealed pong via other misidentified")
  }
}

func TestPongCheck(t *testing.T) {
  // invalid invocation; only applicable to a previous/other relationship
  testHand, testTile := gt.TestHandMaker("🀑🀒🀓🀉🀉🀇🀝🀞🀒🀒🀟🀆🀆🀆;")
  outcome, _ := testHand.HavePong(testTile, "draw")
  if outcome {
    t.Errorf("a pong was misidentified: invalid invocation case")
  }
  
  // previous
  testHand, testTile = gt.TestHandMaker("🀑🀒🀓🀉🀉🀇🀝🀞🀒🀒🀟🀆🀆;🀆")
  outcome, pongTile := testHand.HavePong(testTile, "previous")
  if !outcome || pongTile != "🀆" {
    t.Errorf("a pong was misidentified: previous case")
  }
  
  // other
  testHand, testTile = gt.TestHandMaker("🀑🀒🀓🀉🀉🀇🀝🀞🀒🀒🀟🀆🀆;🀆")
  outcome, pongTile = testHand.HavePong(testTile, "other")
  if !outcome || pongTile != "🀆" {
    t.Errorf("a pong was misidentified: other case")
  }
  
}


