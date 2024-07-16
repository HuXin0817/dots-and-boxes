# Dots and Boxes Game

This project is a Dots and Boxes game developed using the Fyne library in Go. The game includes AI players, a scoring
mechanism, and a graphical representation of the board and game elements. The game allows customization of the board
size, dot distance, AI search time, and other configurations. It also supports saving and loading game states, as well
as performance analysis.

![demo](demo.gif)

## Table of Contents

1. [Features](#features)
2. [Installation](#installation)
3. [Usage](#usage)
4. [Configuration](#configuration)
5. [Game Controls](#game-controls)
6. [AI and Performance Analysis](#ai-and-performance-analysis)
7. [Contributing](#contributing)
8. [License](#license)

## Features

- Customizable board size and dot distance.
- AI players with adjustable search time and goroutines.
- Automatic game restart and music options.
- Performance analysis and profiling.
- Saving and loading game states.
- Graphical representation of the game board using Fyne.
- Menu shortcuts for various game actions.

## Installation

1. **Clone the repository:**

    ```bash
    git clone https://github.com/HuXin0817/dots-and-boxes.git
    cd dots-and-boxes
    ```

2. **Install dependencies:**

    ```bash
    go get -u fyne.io/fyne/v2
    go get -u github.com/bytedance/sonic
    go get -u github.com/gin-gonic/gin
    go get -u github.com/gin-contrib/pprof
    go get -u github.com/hajimehoshi/oto
    go get -u github.com/hajimehoshi/go-mp3
    go get -u github.com/faiface/beep
    ```

3. **Build the project:**

    ```bash
    go build -o dots-and-boxes main.go
    ```

4. **Run the game:**

    ```bash
    ./dots-and-boxes
    ```

## Usage

1. **Starting the game:**
   When you run the game, a window will open with the Dots and Boxes board. The game will start automatically with the
   default configuration.

2. **Interacting with the game:**
    - Click on the edges to add lines and form boxes.
    - The current player's turn is highlighted.
    - Scores are updated as players complete boxes.

3. **Saving and loading game states:**
    - Game states are automatically saved to `meta.json`.
    - The game will load the last saved state when restarted.

## Configuration

The game can be configured through the `ChessMeta` struct in the code. You can adjust various parameters such as:

- `BoardSize`: The size of the board (default is 6).
- `DotCanvasDistance`: The distance between dots (default is 80).
- `AISearchTime`: The time duration for AI search (default is 1 second).
- `AISearchGoroutines`: The number of goroutines for AI search (default is the number of CPU cores).
- `AutoRestartGame`: Flag for auto-restarting the game.
- `OpenMusic`: Flag for playing music during the game.

## Game Controls

- **Restart Game:** Press `R` to restart the game with the current board size.
- **Undo Move:** Press `Z` to undo the last move.
- **Show Scores:** Press `T` to display the current scores.
- **Save Screenshot:** Press `S` to save a screenshot of the game.
- **Adjust Board Width:** Press `Up` to increase and `Down` to decrease the board width.
- **Adjust Board Size:** Press `=` to increase and `-` to decrease the board size.
- **Toggle AI Player 1:** Press `1` to enable or disable AI for Player 1.
- **Toggle AI Player 2:** Press `2` to enable or disable AI for Player 2.
- **Adjust AI Search Time:** Press `3` to increase and `4` to decrease the AI search time.
- **Adjust AI Search Goroutines:** Press `6` to increase and `7` to decrease the number of goroutines for AI search.
- **Toggle Auto Restart Game:** Press `A` to enable or disable automatic game restart.
- **Toggle Music:** Press `P` to enable or disable music.
- **Save Performance Analysis:** Press `F` to save a performance analysis report.
- **Help:** Press `H` to open the help documentation.

## AI and Performance Analysis

The game includes AI players that can be enabled for Player 1 and Player 2. The AI uses a search engine to evaluate the
best move based on the current board state. You can configure the AI search time and the number of goroutines used for
the search.

The game also supports performance analysis using `pprof`. You can generate performance analysis reports to identify
bottlenecks and optimize the game.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request with your changes. Ensure that your code
follows the existing style and includes tests where appropriate.

## License

This project is licensed under the Mulan PSL v2. See the `LICENSE` file for details.
