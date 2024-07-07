# Dots and Boxes

This is a Dots and Boxes game implemented in Go using the Fyne GUI library. The game supports both AI and human players and provides a visual interface to play the game.

![](./demo.png)

## Table of Contents

- [Installation](#installation)
- [Game Rules](#game-rules)
- [Implementation Details](#implementation-details)
- [Contributing](#contributing)
- [License](#license)

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/HuXin0817/dots-and-boxes.git
   cd dots-and-boxes
   ```

2. **Install dependencies:**

   Make sure you have Go installed on your system. Then, install the required dependencies:

   ```bash
   go mod tidy
   ```

3. **Run the project:**

   ```bash
   go run main.go
   ```

The game window will open, and you can start playing the game.

## Game Rules

Dots and Boxes is a classic game where players take turns to draw lines between dots on a grid. The goal is to complete boxes. Each completed box earns the player an additional turn. The player with the most completed boxes at the end of the game wins.

- **Turn:** Players alternate turns. A turn consists of drawing a line between two adjacent dots.
- **Scoring:** Completing a box earns the player a point and an extra turn.
- **End Game:** The game ends when all possible lines are drawn. The player with the most points wins.

## Implementation Details

### Board Representation

The board is represented using various types:

- **Dot:** Represents a dot on the board.
- **Edge:** Represents a line between two dots.
- **Box:** Represents a box formed by four edges.
- **Board:** A map of edges used to keep track of drawn lines.

### Main Components

- **Turn Management:** The game alternates turns between two players, `Player1` and `Player2`.
- **AI Players:** The game can be configured to have AI players for both `Player1` and `Player2`.
- **Visual Interface:** The game uses the Fyne library to create a graphical user interface, displaying dots, lines, and boxes.

### Game Logic

- **Edge Addition:** When an edge is added, the game checks if any boxes are completed. If a box is completed, the player earns a point and gets another turn.
- **AI Logic:** The AI determines the next move by simulating possible moves and choosing the one that maximizes its advantage while minimizing the opponent's advantage.

### Logging

The game logs important events such as turn changes, edge additions, and game results using the `colog` package. The logs are saved in a `gamelog` directory with a timestamped filename.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

1. **Fork the repository**
2. **Create a new branch** (`git checkout -b feature-branch`)
3. **Commit your changes** (`git commit -am 'Add new feature'`)
4. **Push to the branch** (`git push origin feature-branch`)
5. **Create a new Pull Request**

## License

This project is licensed under the Mulan PSL v2 License - see the [LICENSE](LICENSE) file for details.

---

Enjoy playing Dots and Boxes! If you have any questions or need further assistance, feel free to reach out.
