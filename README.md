## Dots and Boxes Game

This is a Fyne-based implementation of the classic game Dots and Boxes. The game includes various features such as AI
players, score tracking, custom board sizes, and more. Below is a comprehensive guide to help you understand and use the
code.

![demo](demo.gif)

### Features

- **Custom Board Size**: Adjust the size of the game board.
- **AI Players**: Play against AI or let AI play against each other.
- **Score Tracking**: Keep track of player scores.
- **Music**: Background music during gameplay.
- **Performance Analysis**: Measure and analyze the performance of the game.

### Requirements

- Go programming language (>=1.16)
- Fyne library for GUI
- Sonic library for JSON operations
- Beep library for audio playback
- Gin and pprof for performance analysis

### Installation

1. Install Go from the [official website](https://golang.org/dl/).
2. Install the required libraries:
   ```sh
   go get fyne.io/fyne/v2
   go get github.com/bytedance/sonic
   go get github.com/faiface/beep
   go get github.com/faiface/beep/mp3
   go get github.com/gin-gonic/gin
   go get github.com/gin-contrib/pprof
   ```

### Running the Game

1. Clone the repository:
   ```sh
   git clone https://github.com/HuXin0817/dots-and-boxes.git
   cd dots-and-boxes
   ```
2. Build and run the game:
   ```sh
   go run main.go
   ```

### Gameplay

1. **Starting a Game**: The game starts automatically when you run the application. The default board size is 6x6.
2. **Making a Move**: Click on the edges to form a box. If you complete a box, you score a point and get another turn.
3. **Undo Move**: Press 'Z' to undo the last move.
4. **Restart Game**: Press 'R' to restart the game with the current board size.
5. **Adjust Board Size**: Use the menu options to increase or decrease the board size.

### Menu Options

- **Game Menu**:
    - **Restart**: Restart the game.
    - **Undo**: Undo the last move.
    - **Score**: Display current scores.
    - **Quit**: Exit the game.
    - **Help**: Open the help documentation.

- **Board Menu**:
    - **Add Board Width**: Increase the width of the dots.
    - **Reduce Board Width**: Decrease the width of the dots.
    - **Reset Board Width**: Reset the width to default.
    - **Add Board Size**: Increase the board size.
    - **Reduce Board Size**: Decrease the board size.
    - **Reset Board Size**: Reset the board size to default.

- **Config Menu**:
    - **AI Player 1**: Toggle AI for Player 1.
    - **AI Player 2**: Toggle AI for Player 2.
    - **Increase AI Search Time**: Increase the time AI takes to make a move.
    - **Reduce AI Search Time**: Decrease the time AI takes to make a move.
    - **Reset AI Search Time**: Reset AI search time to default.
    - **Increase Search Goroutines**: Increase the number of goroutines for AI search.
    - **Reduce Search Goroutines**: Decrease the number of goroutines for AI search.
    - **Reset Search Goroutines**: Reset search goroutines to default.
    - **Auto Restart**: Toggle auto-restart of the game after completion.
    - **Music**: Toggle background music.

- **Performance Analysis Menu**:
    - **Increase Performance Analysis Time**: Increase the duration for performance analysis.
    - **Reduce Performance Analysis Time**: Decrease the duration for performance analysis.
    - **Reset Performance Analysis Time**: Reset performance analysis time to default.
    - **Save CPU Performance Analysis**: Save the CPU performance analysis data.

### Customization

You can customize the colors, AI logic, and other game parameters by modifying the constants and functions in the code.
For example, to change the colors, update the `LightThemeDotCanvasColor`, `DarkThemeDotCanvasColor`, and other color
variables.

### Troubleshooting

- Ensure all dependencies are installed correctly.
- Make sure you are using a compatible version of Go.
- Check for any error messages in the console and resolve them.

### Contribution

Feel free to fork the repository, make improvements, and submit pull requests. Contributions are welcome!

### License

This project is licensed under the Mulan PSL v2 License.

---

For detailed documentation and updates, visit the [GitHub repository](https://github.com/HuXin0817/dots-and-boxes).