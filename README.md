# Dots and Boxes Game

This project implements the classic Dots and Boxes game using the Fyne library in Go. The game features a graphical user
interface and allows for both human and AI players.

## Features

- Graphical user interface using the Fyne library
- Support for AI players
- Adjustable board size
- Game state logging
- Key bindings for game controls

## Requirements

- Go 1.16 or later
- Fyne v2.0.0 or later
- [colog](https://github.com/HuXin0817/colog) library for logging

## Installation

1. Install Go from [golang.org](https://golang.org/dl/).
2. Set up Go environment variables as described in the Go [installation guide](https://golang.org/doc/install).
3. Install Fyne:

```sh
go get fyne.io/fyne/v2
```

4. Install the colog library:

```sh
go get github.com/HuXin0817/colog
```

5. Clone the repository:

```sh
git clone https://github.com/HuXin0817/dots-and-boxes.git
cd dots-and-boxes
```

## Running the Game

To run the game, use the following command:

```sh
go run main.go
```

## Controls

- **R**: Reset the game
- **1**: Toggle AI player 1
- **2**: Toggle AI player 2
- **+**/**=**/**Up Arrow**: Increase board size
- **-**/**Down Arrow**: Decrease board size
- **W**: Set board size to 6x6
- **Q**: Quit the game
- **Space**: Pause/Continue the game
- **L**: Toggle game state logging

## Game Structure

The game is implemented using several key structs and types:

- **Game**: Manages the game state, including the board, player scores, and turn management.
- **Board**: Represents the game board and manages the edges and boxes.
- **Dot**: Represents a dot on the board.
- **Edge**: Represents an edge between two dots.
- **Box**: Represents a box formed by four edges.

### AI Player

The AI player selects the best edge to add by evaluating possible moves and using a scoring algorithm to maximize its
chances of winning.

### Logging

The game state can be logged to a file for debugging and analysis. Logging can be toggled on and off using the **L**
key.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.
