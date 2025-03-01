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

function getMovePositionToward(gameState, currentUnit, targetPosition) {
  const unitActions = getAvailableActions(gameState, currentUnit.type);
  if (!unitActions || !unitActions.move || !unitActions.move.distance) {
    return currentUnit.position;
  }

  const maxDistance = unitActions.move.distance;
  const deltaX = targetPosition.x - currentUnit.position.x;
  const deltaY = targetPosition.y - currentUnit.position.y;

  // Calculate euclidean distance to the target
  const totalDistance = calculateEuclideanDistance(
    currentUnit.position,
    targetPosition,
  );

  // If we can reach the target in one move, return the target position
  if (totalDistance <= maxDistance) {
    return {
      x: targetPosition.x,
      y: targetPosition.y,
    };
  }

  // Otherwise, move as far as possible along the line to the target
  const ratio = maxDistance / totalDistance;

  // Use ceiling when moving upward/leftward (negative deltas)
  // Use floor when moving downward/rightward (positive deltas)
  return {
    x:
      deltaX < 0
        ? Math.ceil(currentUnit.position.x + deltaX * ratio)
        : Math.floor(currentUnit.position.x + deltaX * ratio),
    y:
      deltaY < 0
        ? Math.ceil(currentUnit.position.y + deltaY * ratio)
        : Math.floor(currentUnit.position.y + deltaY * ratio),
  };
}

function isValidPosition(gameState, position) {
  if (!gameState || !gameState.map) return false;

  return (
    position.x >= 0 &&
    position.x < gameState.map.width &&
    position.y >= 0 &&
    position.y < gameState.map.height
  );
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
<generated></generated>;
