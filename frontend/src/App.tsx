import { useEffect, useState } from "react";
import logo from "./logo.svg";
import "./App.css";

function Square(props: { value: string; onClick: () => void }) {
  return (
    <button className="square" onClick={() => props.onClick()}>
      {props.value}
    </button>
  );
}

function App() {
  const [board, setBoard] = useState(Array(9).fill(""));

  useEffect(() => {}, []);

  let renderSquare = (i: number) => {
    return (
      <Square
        value={board[i]}
        onClick={() => {
          setBoard(board.map((_, j) => (j === i ? "X" : board[j])));
        }}
      />
    );
  };

  return (
    <div className="game">
      <div className="game-board">
        <div>
          <div className="board-row">
            {renderSquare(0)}
            {renderSquare(1)}
            {renderSquare(2)}
          </div>
          <div className="board-row">
            {renderSquare(3)}
            {renderSquare(4)}
            {renderSquare(5)}
          </div>
          <div className="board-row">
            {renderSquare(6)}
            {renderSquare(7)}
            {renderSquare(8)}
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
