import { getMovePositionToward } from "./js.js";

function calculateEuclideanDistance(point1, point2) {
  const dx = point1.x - point2.x;
  const dy = point1.y - point2.y;
  return Math.hypot(dx, dy); // Using hypot for better numerical stability
}

// Test getMovePositionToward function
function testGetMovePositionToward() {
  // Define reusable map and unit_action_map
  const testMap = { width: 50, height: 50 };
  const testUnitActionMap = {
    soldier: {
      move: { distance: 2 },
    },
  };

  // Test case 1: Moving toward a target within range
  const gameState1 = {
    ...testMap,
    units: [
      { id: 1, position: { x: 10, y: 10 }, type: "soldier" },
      { id: 2, position: { x: 15, y: 15 }, type: "enemy" },
    ],
    unit_action_map: testUnitActionMap,
  };

  const result1 = getMovePositionToward(gameState1, gameState1.units[0], {
    x: 15,
    y: 15,
  });
  const distance1 = calculateEuclideanDistance(
    gameState1.units[0].position,
    result1,
  );
  console.assert(
    distance1 <= 2,
    "TEST1: Should move maximum distance toward target " + distance1,
  );
  console.assert(
    result1.x === 11 && result1.y === 11,
    "TEST1: Should move to target when in range: " + JSON.stringify(result1),
  );

  // Test case 2: Moving toward a target beyond range
  const result2 = getMovePositionToward(gameState1, gameState1.units[0], {
    x: 30,
    y: 30,
  });
  console.assert(
    calculateEuclideanDistance(gameState1.units[0].position, result2) <= 2,
    "TEST2: Should move maximum distance toward target",
  );

  // Test case 3: Target position is already occupied
  const gameState2 = {
    ...testMap,
    units: [
      { id: 1, position: { x: 10, y: 10 }, type: "soldier" },
      { id: 2, position: { x: 11, y: 11 }, type: "healer" },
      { id: 3, position: { x: 13, y: 13 }, type: "healer" },
    ],
    unit_action_map: testUnitActionMap,
  };

  const result3 = getMovePositionToward(gameState2, gameState2.units[0], {
    x: 13,
    y: 13,
  });
  console.assert(
    result3 != null && result3.x === 12 && result3.y === 10,
    "TEST3: Should not move to an occupied position: " +
      JSON.stringify(result3),
  );

  const gameState3 = {
    ...testMap,
    units: [
      { id: 1, position: { x: 10, y: 10 }, type: "soldier" },
      { id: 3, position: { x: 13, y: 13 }, type: "healer" },
    ],
    unit_action_map: {
      soldier: {
        move: { distance: 5 },
      },
    },
  };

  const result4 = getMovePositionToward(gameState3, gameState3.units[0], {
    x: 13,
    y: 13,
  });
  console.assert(
    result4 != null && result4.x !== 13 && result4.y !== 13,
    "TEST3: Should not move to an occupied position: " +
      JSON.stringify(result4),
  );

  console.log("All getMovePositionToward tests completed");
}

// Run the test
testGetMovePositionToward();
