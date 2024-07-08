# Dots and Boxes Game

This project is a graphical implementation of the Dots and Boxes game using the Fyne library in Go. The game includes AI players and a scoring mechanism, with a graphical representation of the board and game elements.

![demo](demo.png)

## Features

- Two players: Player1 and Player2
- AI support for both players
- Interactive graphical user interface
- Animation for scoring boxes
- Game logging



## Prerequisites

- Go 1.16 or later
- Fyne library

## Installation

1. Install Go from the [official website](https://golang.org/dl/).

2. Install the Fyne library:

    ```bash
    go get fyne.io/fyne/v2
    ```

3. Clone the repository:

    ```bash
    git clone https://github.com/HuXin0817/dots-and-boxes-game.git
    ```

4. Navigate to the project directory:

    ```bash
    cd dots-and-boxes-game
    ```

## Usage

1. Run the game:

    ```bash
    go run main.go
    ```

2. The game window will open, and the game will start automatically. The AI will make the first move if enabled.

## Game Rules

- Players take turns to draw edges between dots.
- Completing the fourth edge of a box earns the player a point, and they get another turn.
- The game ends when all boxes are completed.
- The player with the most points wins.

### Board Representation

The board is represented using various types:

- **Dot:** Represents a dot on the board.
- **Edge:** Represents a line between two dots.
- **Box:** Represents a box formed by four edges.
- **Board:** A map of edges used to keep track of drawn lines.

## Customization

- The AI for each player can be enabled or disabled by setting the `AIPlayer1` and `AIPlayer2` constants.
- The game board size can be adjusted by changing the `BoardSize` constant.
- Various visual elements such as colors and animation steps can be customized in the code.

## Logging

- Game actions are logged to a file in the `game log` directory.
- The log file is named with the current date and time.

## Project Structure

- `main.go`: The main game logic and UI implementation.
- Various helper functions and constants are defined for managing the game state, drawing the UI, and handling AI moves.

## License

This project is licensed under the Mulan PSL v2. See the LICENSE file for details.

## Acknowledgments

- The [Fyne](https://fyne.io/) library for providing the GUI framework.
- [colog](https://github.com/HuXin0817/colog) for logging support.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

1. **Fork the repository**
2. **Create a new branch** (`git checkout -b feature-branch`)
3. **Commit your changes** (`git commit -am 'Add new feature'`)
4. **Push to the branch** (`git push origin feature-branch`)
5. **Create a new Pull Request**

## Contact

For any questions or feedback, please contact [202219120810@stu.cdut.edu.cn].

