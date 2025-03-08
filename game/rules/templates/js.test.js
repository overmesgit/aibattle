const test = require("ava");
const {
  getMovePositionToward,
  calculateEuclideanDistance,
} = require("./js.js");

// Define common test fixtures
const testMap = { width: 50, height: 50 };
const testUnitActionMap = {
  soldier: {
    move: { distance: 2 },
  },
};

test("Moving toward a target within range", (t) => {
  // Setup
  const gameState = {
    ...testMap,
    units: [
      { id: 1, position: { x: 10, y: 10 }, type: "soldier" },
      { id: 2, position: { x: 15, y: 15 }, type: "enemy" },
    ],
    unit_action_map: testUnitActionMap,
  };

  // Execute
  const result = getMovePositionToward(gameState, gameState.units[0], {
    x: 15,
    y: 15,
  });

  t.truthy(result, "Result should not be null");
  // Assert
  const distance = calculateEuclideanDistance(
    gameState.units[0].position,
    result,
  );
  t.true(distance <= 2, "Should move maximum distance toward target");
  t.deepEqual(
    result,
    { x: 11, y: 11 },
    "Should move to specific position when in range",
  );
});

test("Moving toward a target beyond range", (t) => {
  // Setup
  const gameState = {
    ...testMap,
    units: [
      { id: 1, position: { x: 10, y: 10 }, type: "soldier" },
      { id: 2, position: { x: 15, y: 15 }, type: "enemy" },
    ],
    unit_action_map: testUnitActionMap,
  };

  // Execute
  const result = getMovePositionToward(gameState, gameState.units[0], {
    x: 30,
    y: 30,
  });

  // Assert
  const distance = calculateEuclideanDistance(
    gameState.units[0].position,
    result,
  );
  t.true(distance <= 2, "Should move maximum distance toward target");
});

test("Should not move to an occupied position - case 1", (t) => {
  // Setup
  const gameState = {
    ...testMap,
    units: [
      { id: 1, position: { x: 10, y: 10 }, type: "soldier" },
      { id: 2, position: { x: 11, y: 11 }, type: "healer" },
      { id: 3, position: { x: 13, y: 13 }, type: "healer" },
    ],
    unit_action_map: testUnitActionMap,
  };

  // Execute
  const result = getMovePositionToward(gameState, gameState.units[0], {
    x: 13,
    y: 13,
  });

  // Assert
  t.truthy(result, "Result should not be null");
  t.deepEqual(result, { x: 12, y: 10 }, "Should find an alternative position");
});

test("Should not move to an occupied position - case 2", (t) => {
  // Setup
  const gameState = {
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

  // Execute
  const result = getMovePositionToward(gameState, gameState.units[0], {
    x: 13,
    y: 13,
  });

  // Assert
  t.truthy(result, "Result should not be null");
  t.notDeepEqual(
    result,
    { x: 13, y: 13 },
    "Should not move to occupied target position",
  );
});

test("Should not move to an occupied position - case 3", (t) => {
  // Setup
  const gameState = {
    ...testMap,
    units: [
      { id: 1, position: { x: 11, y: 10 }, type: "soldier" },
      { id: 3, position: { x: 11, y: 11 }, type: "healer" },
    ],
    unit_action_map: {
      soldier: {
        move: { distance: 5 },
      },
    },
  };

  // Execute
  const result = getMovePositionToward(gameState, gameState.units[0], {
    x: 11,
    y: 11,
  });

  console.log(result);
  // Assert
  t.falsy(result, "Result should be null");
});

test("Handles empty game state gracefully", (t) => {
  // Setup
  const gameState = null;
  const currentUnit = { position: { x: 10, y: 10 }, type: "soldier" };
  const targetPosition = { x: 15, y: 15 };

  // Execute
  const result = getMovePositionToward(gameState, currentUnit, targetPosition);

  // Assert
  t.falsy(result, "Should return null when game state is invalid");
});

test("Handles undefined target position", (t) => {
  // Setup
  const gameState = {
    ...testMap,
    units: [{ id: 1, position: { x: 10, y: 10 }, type: "soldier" }],
    unit_action_map: testUnitActionMap,
  };

  // Execute
  const result = getMovePositionToward(
    gameState,
    gameState.units[0],
    undefined,
  );

  // Assert
  t.falsy(result, "Should return null for undefined target");
});

test("Handles target position outside map boundaries", (t) => {
  // Setup
  const gameState = {
    ...testMap,
    units: [{ id: 1, position: { x: 10, y: 10 }, type: "soldier" }],
    unit_action_map: testUnitActionMap,
  };

  // Execute
  const result = getMovePositionToward(gameState, gameState.units[0], {
    x: -5,
    y: -5,
  });

  // Assert
  t.falsy(result, "Should return null for out-of-bounds target");
});

test("Handles unit with no movement ability", (t) => {
  // Setup
  const gameState = {
    ...testMap,
    units: [{ id: 1, position: { x: 10, y: 10 }, type: "immobile" }],
    unit_action_map: {
      immobile: {}, // No move action
    },
  };

  // Execute
  const result = getMovePositionToward(gameState, gameState.units[0], {
    x: 15,
    y: 15,
  });

  // Assert
  t.falsy(result, "Should return null when unit can't move");
});

test("Moving around obstacles", (t) => {
  // Setup - create a map with units blocking direct path
  const gameState = {
    ...testMap,
    units: [
      { id: 1, position: { x: 10, y: 10 }, type: "soldier" },
      { id: 2, position: { x: 11, y: 11 }, type: "obstacle" },
      { id: 3, position: { x: 12, y: 12 }, type: "obstacle" },
      { id: 4, position: { x: 15, y: 15 }, type: "target" },
    ],
    unit_action_map: {
      soldier: {
        move: { distance: 3 },
      },
    },
  };

  // Execute
  const result = getMovePositionToward(gameState, gameState.units[0], {
    x: 15,
    y: 15,
  });

  // Assert
  t.truthy(result, "Should find a path around obstacles");
  t.notDeepEqual(
    result,
    { x: 11, y: 11 },
    "Should not move to occupied position",
  );
  t.notDeepEqual(
    result,
    { x: 12, y: 12 },
    "Should not move to occupied position",
  );

  const distance = calculateEuclideanDistance(
    gameState.units[0].position,
    result,
  );
  t.true(distance <= 3, "Should move within the maximum range");
});

test("Moving to current position", (t) => {
  // Setup
  const gameState = {
    ...testMap,
    units: [{ id: 1, position: { x: 10, y: 10 }, type: "soldier" }],
    unit_action_map: testUnitActionMap,
  };

  // Execute
  const result = getMovePositionToward(gameState, gameState.units[0], {
    x: 10,
    y: 10,
  });

  // Assert
  t.falsy(result, "Should return null when target is current position");
});

test("getCurrentUnit retrieves unit by ID", (t) => {
  // Setup
  const gameState = {
    units: [
      { id: 1, position: { x: 10, y: 10 }, type: "soldier" },
      { id: 2, position: { x: 15, y: 15 }, type: "healer" },
    ],
  };

  // Execute
  const { getCurrentUnit } = require("./js.js");
  const result = getCurrentUnit(gameState, 2);

  // Assert
  t.deepEqual(result, gameState.units[1], "Should return correct unit");
});

test("getFriendlyUnits returns units on same team", (t) => {
  // Setup
  const gameState = {
    units: [
      { id: 1, team: "A", position: { x: 10, y: 10 }, type: "soldier" },
      { id: 2, team: "A", position: { x: 15, y: 15 }, type: "healer" },
      { id: 3, team: "B", position: { x: 20, y: 20 }, type: "enemy" },
    ],
  };

  // Execute
  const { getFriendlyUnits } = require("./js.js");
  const result = getFriendlyUnits(gameState, 1);

  // Assert
  t.deepEqual(
    result,
    [gameState.units[1]],
    "Should return only friendly units",
  );
});

test("getEnemyUnits returns units on different team", (t) => {
  // Setup
  const gameState = {
    units: [
      { id: 1, team: "A", position: { x: 10, y: 10 }, type: "soldier" },
      { id: 2, team: "A", position: { x: 15, y: 15 }, type: "healer" },
      { id: 3, team: "B", position: { x: 20, y: 20 }, type: "enemy" },
    ],
  };

  // Execute
  const { getEnemyUnits } = require("./js.js");
  const result = getEnemyUnits(gameState, 1);

  // Assert
  t.deepEqual(result, [gameState.units[2]], "Should return only enemy units");
});

test("findNearestEnemy returns closest enemy unit", (t) => {
  // Setup
  const currentUnit = { position: { x: 10, y: 10 } };
  const enemyUnits = [
    { position: { x: 20, y: 20 } },
    { position: { x: 15, y: 15 } },
    { position: { x: 30, y: 30 } },
  ];

  // Execute
  const { findNearestEnemy } = require("./js.js");
  const result = findNearestEnemy(currentUnit, enemyUnits);

  // Assert
  t.deepEqual(result, enemyUnits[1], "Should return closest enemy");
});

test("canAttack validates attack range", (t) => {
  // Setup
  const gameState = {
    unit_action_map: {
      soldier: {
        attack: { range: 5 },
      },
    },
  };
  const attacker = { type: "soldier", position: { x: 10, y: 10 } };
  const targetInRange = { position: { x: 13, y: 13 } };
  const targetOutOfRange = { position: { x: 20, y: 20 } };

  // Execute
  const { canAttack } = require("./js.js");
  const canAttackInRange = canAttack(
    gameState,
    attacker,
    targetInRange,
    "attack",
  );
  const canAttackOutOfRange = canAttack(
    gameState,
    attacker,
    targetOutOfRange,
    "attack",
  );

  // Assert
  t.true(canAttackInRange, "Should allow attack when target is in range");
  t.false(
    canAttackOutOfRange,
    "Should prevent attack when target is out of range",
  );
});

test("isValidPosition validates map boundaries and occupancy", (t) => {
  // Setup
  const gameState = {
    width: 50,
    height: 50,
    units: [{ position: { x: 25, y: 25 } }],
  };

  // Execute
  const { isValidPosition } = require("./js.js");

  // Assert
  t.true(
    isValidPosition(gameState, { x: 10, y: 10 }),
    "Should accept valid position",
  );
  t.false(
    isValidPosition(gameState, { x: -1, y: 10 }),
    "Should reject out-of-bounds position",
  );
  t.false(
    isValidPosition(gameState, { x: 25, y: 25 }),
    "Should reject occupied position",
  );
  t.true(
    isValidPosition(gameState, { x: 25, y: 25 }, true),
    "Should accept occupied position with allowOccupied flag",
  );
});

test("MinHeap implements priority queue correctly", (t) => {
  // Setup
  const { MinHeap } = require("./js.js");
  const heap = new MinHeap();

  // Execute & Assert
  t.true(heap.empty(), "New heap should be empty");

  heap.insert("task1", 5);
  heap.insert("task2", 2);
  heap.insert("task3", 7);

  t.false(heap.empty(), "Heap with items should not be empty");
  t.is(heap.pop(), "task2", "Should extract minimum priority item");
  t.is(heap.pop(), "task1", "Should extract next minimum priority item");
  t.is(heap.pop(), "task3", "Should extract last item");
  t.true(heap.empty(), "Heap should be empty after extracting all items");
});
