# Dots and Boxes Game

## Overview

This project implements the classic Dots and Boxes game using the Fyne UI framework in Go. The game includes AI players
and a scoring mechanism, with a graphical representation of the board and game elements. This README will guide you
through the setup, execution, and understanding of the game's code and features.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Running the Game](#running-the-game)
- [Game Rules](#game-rules)
- [Code Structure](#code-structure)
    - [Main Components](#main-components)
    - [Game Logic](#game-logic)
    - [User Interface](#user-interface)
    - [AI Implementation](#ai-implementation)
- [Customization](#customization)
- [Logging and Assessment](#logging-and-assessment)
- [Conclusion](#conclusion)

![demo](./demo.png)

## Prerequisites

Before you begin, ensure you have the following installed on your system:

- [Go](https://go.dev) programming language (version 1.18 or higher)
- [Fyne](https://fyne.io) UI library (v2.0 or higher)

## Installation

1. **Clone the repository:**

   ```sh
   git clone https://github.com/HuXin0817/dots-and-boxes.git
   cd dots-and-boxes
   ```

2. **Install dependencies:**

   Ensure you have the required dependencies by running:

   ```sh
   go mod tidy
   ```

## Running the Game

To run the game, simply execute:

```sh
go run main.go
```

Optionally, you can specify an assessment file:

```sh
go run main.go assess.json
```

## Game Rules

- Players take turns to connect dots with a line.
- Completing the fourth side of a box earns the player a point and an extra turn.
- The game ends when all possible lines are drawn.
- The player with the most completed boxes wins.

## Code Structure

### Main Components

The main components of the game are:

1. **Dot and Box Structures:** Represent the dots and boxes on the board.
2. **Edge Structure:** Represents the connections between dots.
3. **Board:** Represents the current state of the game board.
4. **AI Logic:** Handles the decision-making for AI players.
5. **User Interface:** Manages the graphical representation using Fyne.

### Game Logic

- **Turn Management:** The game alternates turns between Player 1 and Player 2. Each player can be either human or AI.
- **Scoring:** Players score by completing boxes. The current score is updated and displayed in real-time.
- **Edge and Box Management:** The game keeps track of the edges drawn and the boxes completed.

### User Interface

The UI is built using Fyne. Key elements include:

- **Dots and Edges:** Drawn using Fyne's `canvas.Circle` and `canvas.Line`.
- **Boxes:** Represented using `canvas.Rectangle`, which changes color when completed.
- **Buttons:** Allow players to draw edges by clicking.

### AI Implementation

The AI uses a simple strategy to choose edges that maximize its score while minimizing the opponent's potential score.
The AI logic involves:

- **Edge Selection:** AI chooses the best edge to draw based on the current board state.
- **Score Assessment:** AI evaluates the potential score for each possible move.
- **Concurrency:** The AI calculations run in parallel to optimize performance.

## Customization

You can customize various aspects of the game:

1. **Board Size:** Modify the `BoardSize` constant to change the size of the board.
2. **AI Difficulty:** Adjust the `SearchTime` constant to control the AI's search depth.
3. **Colors:** Change the predefined color variables to customize the game's appearance.

## Logging and Assessment

The game logs moves and scores using the `colog` package. Logs are saved in the `gamelog` directory. The game also
supports assessment data to improve AI performance over time:

- **Assessment Table:** Stores evaluation data for different board states.
- **Assessment File:** Specify a custom assessment file to load previous data.

### Example of the logging output

Each move is logged with detailed information including the current step, the player making the move, the chosen edge,
and the scores:

```
2024-07-03 10:51:42 GAME START!
2024-07-03 10:51:48 Step: 0 Player1 Edge: (4, 4) => (5, 4) Player1 Score: 0, Player2 Score: 0
2024-07-03 10:51:55 Step: 1 Player2 Edge: (4, 3) => (5, 3) Player1 Score: 0, Player2 Score: 0
```

## Conclusion

This project provides a fully functional implementation of the Dots and Boxes game with AI support and a graphical
interface using Go and the Fyne library. The game is customizable, and the AI can improve over time using the assessment
data. Whether you're looking to play a quick game against the computer or explore the code to understand its inner
workings, this project offers a solid foundation. Feel free to contribute, suggest improvements, or customize the game
to suit your preferences.

Happy gaming!