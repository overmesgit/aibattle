/**
 * Gets the unit with the given ID from the game state
 * @param {Object} gameState - Current game state
 * @param {number} currentUnitID - ID of the unit to find
 * @returns {Object|null} - The found unit or null
 */
function getCurrentUnit(gameState, currentUnitID) {
  return gameState?.units?.find((unit) => unit.id === currentUnitID) ?? null;
}

/**
 * Gets all friendly units from the same team as the current unit
 * @param {Object} gameState - Current game state
 * @param {number} currentUnitID - ID of the current unit
 * @returns {Array} - Array of friendly units
 */
function getFriendlyUnits(gameState, currentUnitID) {
  const currentUnit = getCurrentUnit(gameState, currentUnitID);
  if (!currentUnit) return [];

  return gameState.units.filter(
    (unit) => unit.team === currentUnit.team && unit.id !== currentUnitID,
  );
}

/**
 * Gets all enemy units (units not on the same team as the current unit)
 * @param {Object} gameState - Current game state
 * @param {number} currentUnitID - ID of the current unit
 * @returns {Array} - Array of enemy units
 */
function getEnemyUnits(gameState, currentUnitID) {
  const currentUnit = getCurrentUnit(gameState, currentUnitID);
  if (!currentUnit) return [];

  return gameState.units.filter((unit) => unit.team !== currentUnit.team);
}

/**
 * Calculates the Euclidean distance between two points
 * @param {Object} point1 - First point {x, y}
 * @param {Object} point2 - Second point {x, y}
 * @returns {number} - The distance between the points
 */
function calculateEuclideanDistance(point1, point2) {
  const dx = point1.x - point2.x;
  const dy = point1.y - point2.y;
  return Math.hypot(dx, dy); // Using hypot for better numerical stability
}

/**
 * Gets the available actions for a unit type
 * @param {Object} gameState - Current game state
 * @param {string} unitType - Type of the unit
 * @returns {Object} - Map of available actions
 */
function getAvailableActions(gameState, unitType) {
  return gameState?.unit_action_map?.[unitType] ?? {};
}

/**
 * Finds the nearest enemy to a unit
 * @param {Object} currentUnit - Current unit
 * @param {Array} enemyUnits - Array of enemy units
 * @returns {Object|null} - The nearest enemy or null
 */
function findNearestEnemy(currentUnit, enemyUnits) {
  if (!enemyUnits?.length) return null;

  return (
    enemyUnits.reduce((nearest, enemy) => {
      const distance = calculateEuclideanDistance(
        currentUnit.position,
        enemy.position,
      );

      if (!nearest || distance < nearest.distance) {
        return { unit: enemy, distance };
      }
      return nearest;
    }, null)?.unit ?? null
  );
}

/**
 * Checks if a unit can attack a target with a specific attack type
 * @param {Object} gameState - Current game state
 * @param {Object} attacker - Attacking unit
 * @param {Object} target - Target unit
 * @param {string} attackType - Type of attack
 * @returns {boolean} - Whether the attack is possible
 */
function canAttack(gameState, attacker, target, attackType) {
  const range =
    gameState?.unit_action_map?.[attacker.type]?.[attackType]?.range;
  if (!range) return false;

  const distance = calculateEuclideanDistance(
    attacker.position,
    target.position,
  );

  return distance <= range;
}

/**
 * Priority queue implementation using a binary min heap
 */
class MinHeap {
  constructor() {
    this.heap = [];
  }

  empty() {
    return this.heap.length === 0;
  }

  insert(item, priority) {
    this.heap.push({ item, priority });
    this.bubbleUp(this.heap.length - 1);
  }

  pop() {
    if (this.empty()) return null;

    const min = this.heap[0];
    const end = this.heap.pop();

    if (!this.empty()) {
      this.heap[0] = end;
      this.sinkDown(0);
    }

    return min.item;
  }

  bubbleUp(idx) {
    const item = this.heap[idx];

    while (idx > 0) {
      const parentIdx = Math.floor((idx - 1) / 2);
      const parent = this.heap[parentIdx];

      if (item.priority >= parent.priority) break;

      this.heap[parentIdx] = item;
      this.heap[idx] = parent;
      idx = parentIdx;
    }
  }

  sinkDown(idx) {
    const length = this.heap.length;
    const item = this.heap[idx];

    while (true) {
      const leftChildIdx = 2 * idx + 1;
      const rightChildIdx = 2 * idx + 2;
      let swapIdx = null;

      if (
        leftChildIdx < length &&
        this.heap[leftChildIdx].priority < item.priority
      ) {
        swapIdx = leftChildIdx;
      }

      if (rightChildIdx < length) {
        if (
          (swapIdx === null &&
            this.heap[rightChildIdx].priority < item.priority) ||
          (swapIdx !== null &&
            this.heap[rightChildIdx].priority <
              this.heap[leftChildIdx].priority)
        ) {
          swapIdx = rightChildIdx;
        }
      }

      if (swapIdx === null) break;

      this.heap[idx] = this.heap[swapIdx];
      this.heap[swapIdx] = item;
      idx = swapIdx;
    }
  }

  contains(item, compareFunc) {
    return this.heap.some((node) => compareFunc(node.item, item));
  }

  update(item, priority, compareFunc) {
    const idx = this.heap.findIndex((node) => compareFunc(node.item, item));
    if (idx !== -1) {
      this.heap[idx].priority = priority;
      this.bubbleUp(idx);
    }
  }
}

/**
 * A* pathfinding algorithm implementation
 * @param {Object} gameState - Current game state
 * @param {Object} start - Starting position {x, y}
 * @param {Object} goal - Goal position {x, y}
 * @returns {Array|null} - Path to the goal or null if no path exists
 */
function aStar(gameState, start, goal) {
  const posKey = (p) => `${p.x},${p.y}`;
  const posEquals = (a, b) => a.x === b.x && a.y === b.y;
  if (!isValidPosition(gameState, goal, true)) {
    return null;
  }
  // Special case: start equals goal
  if (posEquals(start, goal)) {
    return [start];
  }

  const openSet = new MinHeap();
  const closed = new Map();
  const gScore = new Map();
  const fScore = new Map();
  const from = new Map();

  const startKey = posKey(start);
  gScore.set(startKey, 0);
  fScore.set(startKey, calculateEuclideanDistance(start, goal));
  openSet.insert(start, fScore.get(startKey));

  const dirs = [
    [1, 0],
    [0, 1],
    [-1, 0],
    [0, -1],
  ];

  const getPath = (current, currentKey) => {
    const path = [current];
    let nodeKey = currentKey;

    while (from.has(nodeKey)) {
      path.unshift(from.get(nodeKey));
      nodeKey = posKey(from.get(nodeKey));
    }

    return path;
  };

  // Add iteration limit to prevent infinite loops
  const maxIterations = 100;
  let iterations = 0;
  let current;
  let currentKey;
  while (!openSet.empty() && iterations < maxIterations) {
    iterations++;
    current = openSet.pop();
    currentKey = posKey(current);

    // Goal reached
    if (current.x === goal.x && current.y === goal.y) {
      return getPath(current, currentKey);
    }

    closed.set(currentKey, true);

    for (const [dx, dy] of dirs) {
      const neighbor = { x: current.x + dx, y: current.y + dy };
      const neighborKey = posKey(neighbor);

      // Check if this is the goal position to handle special case
      const isGoalPosition = neighbor.x === goal.x && neighbor.y === goal.y;

      // For non-goal positions, check normal validity; for goal, only check map boundaries
      const isPositionValid = isValidPosition(
        gameState,
        neighbor,
        isGoalPosition,
      );

      if (!isPositionValid || closed.has(neighborKey)) {
        continue;
      }

      // Uniform cost for cardinal directions
      const tentativeG = gScore.get(currentKey) + 1;

      if (tentativeG < (gScore.get(neighborKey) ?? Infinity)) {
        from.set(neighborKey, current);
        gScore.set(neighborKey, tentativeG);
        const f = tentativeG + calculateEuclideanDistance(neighbor, goal);
        fScore.set(neighborKey, f);

        if (openSet.contains(neighbor, posEquals)) {
          openSet.update(neighbor, f, posEquals);
        } else {
          openSet.insert(neighbor, f);
        }
      }
    }
  }

  return getPath(current, currentKey);
}

/**
 * Calculates the furthest position a unit can move towards a target
 * @param {Object} gameState - Current game state
 * @param {Object} currentUnit - Unit that will move
 * @param {Object} targetPosition - Destination position {x, y}
 * @returns {Object|null} - Best move position or null if no move possible
 */
function getMovePositionToward(gameState, currentUnit, targetPosition) {
  const actions = getAvailableActions(gameState, currentUnit.type);
  const path = aStar(gameState, currentUnit.position, targetPosition);

  if (!path?.length) return null;

  const moveDistance = actions.move?.distance ?? 0;
  let furthestReachable = null;

  // Find the furthest position within movement range
  for (let i = 0; i < path.length; i++) {
    const distance = calculateEuclideanDistance(currentUnit.position, path[i]);
    if (distance <= moveDistance) {
      furthestReachable = path[i];
    } else {
      break;
    }
  }

  return furthestReachable;
}

/**
 * Checks if a position is valid (within map bounds and not occupied)
 * @param {Object} gameState - Current game state
 * @param {Object} position - Position to check {x, y}
 * @param allowOccupied - allowed position to be occupied by other unit
 * @returns {boolean} - Whether the position is valid
 */
function isValidPosition(gameState, position, allowOccupied = false) {
  if (!gameState) return false;

  // Check map boundaries
  if (
    position.x < 0 ||
    position.x >= gameState.width ||
    position.y < 0 ||
    position.y >= gameState.height
  ) {
    return false;
  }

  if (allowOccupied) return true;

  // Check unit occupancy
  return !gameState.units?.some(
    (unit) => unit.position.x === position.x && unit.position.y === position.y,
  );
}

/**
 You must implement function GetTurnActions(gameState, currentUnitID, actionIndex) at generated tag
 You can only use js code that can be run in QuickJS.
 QuickJS supports the ES2023 specification

 * Determines the next action for a unit based on the game state and action index
 * @param {Object} gameState - The current game state containing map, units, and rules
 * @param {number} currentUnitID - The ID of the unit taking the action
 * @param {string} actionIndex - The action index ("FirstAction" or "SecondAction")
 * @returns {Object} - A JSON object representing the next action (e.g., {"action":"move","target":{"x":20,"y":20}})
 */
