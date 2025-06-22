import { useState } from "react";

const Signal = (props: { delay: number }) => {
  const { delay } = props;
  const [hovered, setHovered] = useState(false);

  let bars = 0;
  if (delay < 20) bars = 4;
  else if (delay < 50) bars = 3;
  else if (delay < 100) bars = 2;
  else if (delay < 200) bars = 1;
  else bars = 0;

  const barColor = (index: number) => (index < bars ? "#098842" : "#D9D9D9");

  return (
    <div
      style={{ position: "relative", display: "inline-block" }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <svg xmlns="http://www.w3.org/2000/svg" className="w-12" fill="none" viewBox="0 0 523 300">
        <rect width="100" height="300" fill={barColor(0)} rx="24"></rect>
        <rect width="100" height="300" x="141" fill={barColor(1)} rx="24"></rect>
        <rect width="100" height="300" x="282" fill={barColor(2)} rx="24"></rect>
        <rect width="100" height="300" x="423" fill={barColor(3)} rx="24"></rect>
      </svg>
      {hovered && (
        <div
          style={{
            position: "absolute",
            bottom: -30,
            left: "50%",
            transform: "translateX(-50%)",
            padding: "4px 8px",
            backgroundColor: "black",
            color: "white",
            borderRadius: 4,
            fontSize: 12,
            whiteSpace: "nowrap",
            pointerEvents: "none",
            userSelect: "none",
          }}
        >
          {delay} ms
        </div>
      )}
    </div>
  );
};

export default Signal;
