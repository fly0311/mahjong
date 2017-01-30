# mahjong
A computer-mediated Mah Jong game implemented in Go

## Building

`go build main.go`

## Playing

`./main`

Acknowledge the change in player by pressing `[enter]`.

When prompted, enter a valid value (e.g., `y`, `n`, `0-n`), as requested. If an invalid value is provided, `0` is usually assumed.

### Single player mode

`./main -singlePlayer=true`

The computer players in single player mode currently accept presented win opportunities and, naively, accept pong, triple, and seq opportunities, even if unnecessary or strategically suboptimal.

Discard tile selection aims to retain intact sets and preferentially preserves plausible pairs, consecutive tiles that are not at the ends (to allow for up to two matching opportunities), consecutive tiles at the ends, and gapped consecutive tiles.

### Log gameplay actions

`./main -logFile=[filepath]`

`tail -f [filepath]`

Use the log file to keep track of previous player actions (e.g., when sets were revealed, full discard history), if of interest.

## Info

Additional details at <https://www.0n0e.com/public/mahjong/>.

### Game mechanics

Mah Jong is a tile-based table game for four players (and is not the single-player tile-based matching game). There are 144 tiles and, aside from eight “special tiles” (🀢🀣🀤🀥 🀦🀧🀨🀩) that cannot be used to form sets and only affect scoring, there are four of each tile. There are three suits, each with sequential values from one to nine. There is one additional suit that contains seven non-sequential options (🀀🀁🀂🀃 🀄🀅🀆). The objective is to form a winning hand (a pair and four sets or a pair formed with and the presence of one of each the seven non-sequential tiles and values one and nine from the three suits).

A set can be a set of four identical tiles, a set of three identical tiles, or a sequence of three consecutive tiles of the the same suit.

#### Initialization

1. All 144 tiles are shuffled and formed into four “walls”, one in front of each player, each 18 tiles wide and two tiles tall.
2. At the beginning of a session, someone (player 0) is chosen tentatively to be in the East position. If the session is already in progress, the East position is the player who was previously South unless the player in the East position is the winner of the previous game or if the game ended in a draw. The South, West, and North positions are assigned in counter-clockwise order, following the East position. Note that the player numbers do not change over time.
3. At the beginning of a session, the starting East position is revised based on the tentative East position's dice roll. Once the three dice are rolled, the sum-1, modulo 4, determines who will start as East. The wall and initial deal position are based on the tentative East position, rather than the revised East. If a session is already in progress, the current player in the East position (position 0) rolls three dice. The sum-1, modulo 4, determines the “wall” from which dealing will begin. Note that position 0 is always tied to the East position and East changes as the game progresses. For example, if 12 is the sum, dealing will begin from the North position's “wall” (and, if at the start of a session, North will become East at this time).
4. Once the wall is identified, the first block of four tiles (two columns) at the (sum+1)th column, counting from the designated position's right, is dealt to the player in the East position. The next pair of columns (four tiles), to the left, is dealt to the player in the South position. Similarly, for players in the West and North positions. If a “wall”'s end is reached during this process, the next wall, counter-clockwise order, is used for dealing, beginning from the right.
5. The previous process is repeated twice to yield 12 tiles per hand. Conceptually, given perfect shuffling, simply assigning the number of necessary tiles from the start position should be should be identical to this dealing process. However, for possible future presentation purposes, tradition is followed.
6. The top tile from the next available column and the top tile from the third available column are dealt to player in the East position. The player in the South position receives the bottom tile from the first partial column. The player in the West position receives the top tile from the next full column. The player in the North position receives the bottom tile from the now partial column. New tiles are to be drawn starting from the remaining partial column. Dealing should now be complete. The player in the East position should have 14 tiles while all other players should have 13.
7. In order, from the player in the East position, hands are checked for “special tiles”. If one is found, it is revealed and replaced with a replacement tile drawn from the (sum)th column. If the replacement is a “special tile”, the process is repeated (i.e., the replacement “special tile” is made revealed and another replacement tile is drawn). Replacements for “special tiles” and sets of four are drawn from this location. The number of tiles in each hand should not change as a result of this process.

#### Steady-state

There are two phases to the game. The first, informally termed “draw processing”, is where the player operates on a tile added to their hand, typically to determine a tile to discard (from their hand). The second, informally termed “discard processing”, is where other players, not necessarily, the next player, can potentially make use of the discarded tile to either win or reveal a set formed using the discarded tile. If a player reveals a qualifying set, they become the current player and enter “draw processing”.

Refer to the process flow diagram. Note that the players are defined in reference to *i*, which represents the current player.
