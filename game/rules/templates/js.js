function getCurrentUnit(gameState, currentUnitID) {
  if (!gameState || !gameState.units) return null;

  return gameState.units.find((unit) => unit.id === currentUnitID) || null;
}

function getFriendlyUnits(gameState, currentUnitID) {
  const currentUnit = getCurrentUnit(gameState, currentUnitID);
  if (!currentUnit) return [];

  return gameState.units.filter(
    (unit) => unit.team === currentUnit.team && unit.id !== currentUnitID,
  );
}

function getEnemyUnits(gameState, currentUnitID) {
  const currentUnit = getCurrentUnit(gameState, currentUnitID);
  if (!currentUnit) return [];

  return gameState.units.filter((unit) => unit.team !== currentUnit.team);
}

function calculateEuclideanDistance(point1, point2) {
  return Math.sqrt(
    Math.pow(point1.x - point2.x, 2) + Math.pow(point1.y - point2.y, 2),
  );
}

function getAvailableActions(gameState, unitType) {
  if (
    !gameState ||
    !gameState.unit_action_map ||
    !gameState.unit_action_map[unitType]
  ) {
    return {};
  }

  return gameState.unit_action_map[unitType];
}

function findNearestEnemy(currentUnit, enemyUnits) {
  if (!enemyUnits || enemyUnits.length === 0) return null;

  let nearestEnemy = enemyUnits[0];
  let minDistance = calculateEuclideanDistance(
    currentUnit.position,
    nearestEnemy.position,
  );

  for (let i = 1; i < enemyUnits.length; i++) {
    const distance = calculateEuclideanDistance(
      currentUnit.position,
      enemyUnits[i].position,
    );
    if (distance < minDistance) {
      minDistance = distance;
      nearestEnemy = enemyUnits[i];
    }
  }

  return nearestEnemy;
}

function canAttack(gameState, attacker, target, attackType) {
  const unitActions = getAvailableActions(gameState, attacker.type);
  if (
    !unitActions ||
    !unitActions[attackType] ||
    !unitActions[attackType].range
  ) {
    return false;
  }

  const range = unitActions[attackType].range;
  const distance = calculateEuclideanDistance(
    attacker.position,
    target.position,
  );

  return distance <= range;
}

function MinHeap() {
  this.heap = [];

  this.empty = function () {
    return this.heap.length === 0;
  };

  this.insert = function (item, priority) {
    this.heap.push({ item: item, priority: priority });
    this.bubbleUp(this.heap.length - 1);
  };

  this.pop = function () {
    if (this.empty()) return null;

    var min = this.heap[0];
    var end = this.heap.pop();

    if (!this.empty()) {
      this.heap[0] = end;
      this.sinkDown(0);
    }

    return min.item;
  };

  this.bubbleUp = function (idx) {
    var item = this.heap[idx];

    while (idx > 0) {
      var parentIdx = Math.floor((idx - 1) / 2);
      var parent = this.heap[parentIdx];

      if (item.priority >= parent.priority) break;

      this.heap[parentIdx] = item;
      this.heap[idx] = parent;
      idx = parentIdx;
    }
  };

  this.sinkDown = function (idx) {
    var length = this.heap.length;
    var item = this.heap[idx];

    while (true) {
      var leftChildIdx = 2 * idx + 1;
      var rightChildIdx = 2 * idx + 2;
      var swapIdx = null;

      if (leftChildIdx < length) {
        if (this.heap[leftChildIdx].priority < item.priority) {
          swapIdx = leftChildIdx;
        }
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
  };

  this.contains = function (item, compareFunc) {
    for (var i = 0; i < this.heap.length; i++) {
      if (compareFunc(this.heap[i].item, item)) {
        return true;
      }
    }
    return false;
  };

  this.update = function (item, priority, compareFunc) {
    for (var i = 0; i < this.heap.length; i++) {
      if (compareFunc(this.heap[i].item, item)) {
        this.heap[i].priority = priority;
        this.bubbleUp(i);
        return;
      }
    }
  };
}

function aStar(gameState, start, goal) {
  function posKey(p) {
    return p.x + "," + p.y;
  }

  function euclideanDist(a, b) {
    return Math.sqrt(Math.pow(a.x - b.x, 2) + Math.pow(a.y - b.y, 2));
  }

  function posEquals(a, b) {
    return a.x === b.x && a.y === b.y;
  }

  var openSet = new MinHeap();
  var closed = {};
  var gScore = {};
  var fScore = {};
  var from = {};

  var startKey = posKey(start);
  gScore[startKey] = 0;
  fScore[startKey] = euclideanDist(start, goal);
  openSet.insert(start, fScore[startKey]);

  // Directions: including diagonals
  var dirs = [
    [1, 0],
    [0, 1],
    [-1, 0],
    [0, -1],
  ];

  while (!openSet.empty()) {
    var current = openSet.pop();
    var currentKey = posKey(current);

    if (current.x === goal.x && current.y === goal.y) {
      var path = [current];
      var nodeKey = currentKey;

      while (from[nodeKey]) {
        path.unshift(from[nodeKey]);
        nodeKey = posKey(from[nodeKey]);
      }

      return path;
    }

    closed[currentKey] = true;

    for (var i = 0; i < dirs.length; i++) {
      var dx = dirs[i][0],
        dy = dirs[i][1];
      var neighbor = { x: current.x + dx, y: current.y + dy };
      var neighborKey = posKey(neighbor);

      // Special case: if this is the goal position, we can ignore unit occupancy
      var isGoalPosition = neighbor.x === goal.x && neighbor.y === goal.y;

      // For non-goal positions, check normal validity; for goal position, only check map boundaries
      var isPositionValid = isGoalPosition
        ? neighbor.x >= 0 &&
          neighbor.x < gameState.width &&
          neighbor.y >= 0 &&
          neighbor.y < gameState.height
        : isValidPosition(gameState, neighbor);

      if (!isPositionValid || closed[neighborKey]) {
        continue;
      }

      // Cost is 1.0 for cardinal directions, sqrt(2) ≈ 1.414 for diagonals
      var moveCost = dx === 0 || dy === 0 ? 1.0 : Math.SQRT2;
      var tentativeG = gScore[currentKey] + moveCost;

      if (tentativeG < (gScore[neighborKey] || Infinity)) {
        from[neighborKey] = current;
        gScore[neighborKey] = tentativeG;
        fScore[neighborKey] = tentativeG + euclideanDist(neighbor, goal);

        if (openSet.contains(neighbor, posEquals)) {
          openSet.update(neighbor, fScore[neighborKey], posEquals);
        } else {
          openSet.insert(neighbor, fScore[neighborKey]);
        }
      }
    }
  }

  return null;
}

/**
 * Calculates the furthest position a unit can move towards a target
 * @param {Object} gameState - Current game state
 * @param {Object} currentUnit - Unit that will move
 * @param {Object} targetPosition - Destination position {x, y}
 * @returns {Object|null} - Best move position or null if no move possible
 */
function getMovePositionToward(gameState, currentUnit, targetPosition) {
  var actions = getAvailableActions(gameState, currentUnit.type);
  var path = aStar(gameState, currentUnit.position, targetPosition);
  if (!path || path.length === 0) {
    return null;
  }

  var nextPosition = path[0];
  for (var i = 0; i < path.length; i++) {
    if (
      calculateEuclideanDistance(currentUnit.position, path[i]) <
      actions.move.distance
    ) {
      nextPosition = path[i];
    } else {
      return nextPosition;
    }
  }

  return null;
}

function isValidPosition(gameState, position) {
  if (!gameState) return false;

  // Check if position is within map boundaries
  if (
    position.x < 0 ||
    position.x >= gameState.width ||
    position.y < 0 ||
    position.y >= gameState.height
  ) {
    return false;
  }

  // Check if position is already occupied by any unit
  if (gameState.units) {
    for (let i = 0; i < gameState.units.length; i++) {
      const unit = gameState.units[i];
      if (unit.position.x === position.x && unit.position.y === position.y) {
        return false;
      }
    }
  }

  return true;
}

/**
 You must implement function GetTurnActions(gameState, currentUnitID, actionIndex) at generated tag
 You can only use js code that can be run in github.com/dop251/goja
 Goja is an implementation of ECMAScript 5.1 in pure Go with emphasis on standard compliance and performance.

 * Determines the next action for a unit based on the game state and action index
 * @param {Object} gameState - The current game state containing map, units, and rules
 * @param {number} currentUnitID - The ID of the unit taking the action
 * @param {string} actionIndex - The action index ("FirstAction" or "SecondAction")
 * @returns {Object} - A JSON object representing the next action (e.g., {"action":"move","target":{"x":20,"y":20}})
 */
