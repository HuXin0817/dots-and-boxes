# Dots and Boxes Game

This repository contains the implementation of the Dots and Boxes game using the Fyne library in Go. The game features
AI players, scoring mechanisms, and a graphical representation of the board and game elements.

![demo](demo.gif)

## Features

- **Graphical User Interface:** The game uses the Fyne library for a user-friendly and cross-platform graphical
  interface.
- **AI Players:** Options to enable AI for Player 1 and Player 2.
- **Scoring Mechanism:** Real-time score updates for players.
- **Animation:** Smooth animations for visual feedback.
- **Music:** Background music and sound effects for moves and scores.
- **Responsive Board:** The board size and dot distance can be adjusted dynamically.
- **Game Controls:** Various controls and shortcuts for gameplay.

## Installation

To run the game, you need to have Go installed on your system. Follow the steps below to set up the project:

1. **Clone the repository:**
   ```bash
   git clone https://github.com/HuXin0817/dots-and-boxes.git
   cd dots-and-boxes
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the game:**
   ```bash
   go run main.go
   ```

## Gameplay

### Controls

- **Space:** Pause/Resume the game.
- **R:** Restart the game.
- **1:** Toggle AI for Player 1.
- **2:** Toggle AI for Player 2.
- **A:** Toggle auto-restart after a game ends.
- **Up Arrow:** Increase board size.
- **Down Arrow:** Decrease board size.
- **Left Arrow:** Decrease dot distance.
- **Right Arrow:** Increase dot distance.
- **Z:** Undo the last move.
- **W:** Reset board size to 6x6.
- **Q:** Quit the game.
- **M:** Toggle music on/off.

### Options

- **AIPlayer1:** Enable/disable AI for Player 1.
- **AIPlayer2:** Enable/disable AI for Player 2.
- **PauseState:** Pause/resume the game.
- **AutoRestart:** Automatically restart the game after it ends.
- **Music:** Toggle background music and sound effects.

## Code Structure

- **main.go:** The main file containing the game logic and UI implementation.
- **types.go:** Contains type definitions and utility functions.
- **ai.go:** AI implementation and search algorithms.
- **utils.go:** Utility functions for gameplay and animations.
- **theme.go:** Custom theme implementation for the game.
- **music.go:** Background music and sound effects control.

## Customization

You can customize the game by modifying the constants and functions defined in the main.go and utils.go files. Adjust
the board size, dot distance, colors, and other parameters to suit your preferences.

## License

This project is licensed under the Mulan PSL v2 License. See the LICENSE file for more details.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue if you have any suggestions or
find any bugs.

## Contact

For any questions or feedback, you can reach out to the project maintainer at your-email@example.com.

Enjoy the game!