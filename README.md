# Dots-and-Boxes Game

A classic strategy game implemented using Go and Fyne library for the user interface. This project provides an engaging
and interactive experience with customizable settings and AI opponents.

![demo](demo.gif)

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Game Rules](#game-rules)
- [Controls](#controls)
- [Configuration](#configuration)
- [Help](#help)
- [Contributing](#contributing)
- [License](#license)

## Features

- Two-player gameplay.
- AI opponents for both Player 1 and Player 2.
- Customizable board size.
- Undo last move functionality.
- Pause and resume the game.
- Background music with on/off toggle.
- Automatic game restart option.
- Visual and sound effects for moves and scoring.

## Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/HuXin0817/dots-and-boxes.git
   cd dots-and-boxes
   ```

2. **Install dependencies:**
   Ensure you have Go installed. You can download and install it from [here](https://golang.org/dl/).

   Install the Fyne library:
   ```bash
   go get fyne.io/fyne/v2
   ```

3. **Build and run the application:**
   ```bash
   go build -o dots-and-boxes
   ./dots-and-boxes
   ```

## Usage

Run the application and use the mouse to connect dots and form boxes. The game will automatically switch turns between
players or AI.

## Game Rules

- The objective is to form more boxes than your opponent.
- Players take turns to connect two adjacent dots with a horizontal or vertical line.
- Completing the fourth side of a box scores a point and grants an extra turn.
- The game ends when all possible lines are drawn, and the player with the most boxes wins.

## Controls

- **Click** on the edges between dots to draw a line.
- **Menu Options:**
    - **Game:**
        - **Restart:** Start a new game with the current board size.
        - **Undo:** Undo the last move made.
        - **Pause:** Pause the game.
        - **Score:** Display the current score.
        - **Quit:** Exit the game.
    - **Board:**
        - **Add Board Size:** Increase the size of the board.
        - **Reduce Board Size:** Decrease the size of the board.
        - **Reset Board:** Reset the board to default size and settings.
    - **Config:**
        - **AI Player 1:** Toggle AI for Player 1.
        - **AI Player 2:** Toggle AI for Player 2.
        - **Auto Restart:** Automatically restart the game after it ends.
        - **Music:** Toggle background music on/off.
    - **Help:**
        - **Help:** Display the help document.

## Configuration

- Customize the game settings through the menu options for a personalized experience.
- Adjust board size and toggle AI players for different levels of challenge.

## Help

For detailed instructions on gameplay and controls, refer to the in-game help section under the Help menu.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request with your improvements.

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-branch`).
5. Create a new pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.