function getCurrentUnit(gameState, currentUnitID) {
    if (!gameState || !gameState.units) return null;

    return gameState.units.find(unit => unit.id === currentUnitID) || null;
}

function getFriendlyUnits(gameState, currentUnitID) {
    const currentUnit = getCurrentUnit(gameState, currentUnitID);
    if (!currentUnit) return [];

    return gameState.units.filter(unit => unit.team === currentUnit.team && unit.id !== currentUnitID);
}

function getEnemyUnits(gameState, currentUnitID) {
    const currentUnit = getCurrentUnit(gameState, currentUnitID);
    if (!currentUnit) return [];

    return gameState.units.filter(unit => unit.team !== currentUnit.team);
}

function calculateEuclideanDistance(point1, point2) {
    return Math.sqrt(Math.pow(point1.x - point2.x, 2) + Math.pow(point1.y - point2.y, 2));
}

function getAvailableActions(gameState, unitType) {
    if (!gameState || !gameState.rules || !gameState.rules.unitTypes || !gameState.rules.unitTypes[unitType]) {
        return {};
    }

    return gameState.rules.unitTypes[unitType];
}

function findNearestEnemy(currentUnit, enemyUnits) {
    if (!enemyUnits || enemyUnits.length === 0) return null;

    let nearestEnemy = enemyUnits[0];
    let minDistance = calculateManhattanDistance(currentUnit.position, nearestEnemy.position);

    for (let i = 1; i < enemyUnits.length; i++) {
        const distance = calculateManhattanDistance(currentUnit.position, enemyUnits[i].position);
        if (distance < minDistance) {
            minDistance = distance;
            nearestEnemy = enemyUnits[i];
        }
    }

    return nearestEnemy;
}

function canAttack(gameState, attacker, target, attackType) {
    const unitActions = getAvailableActions(gameState, attacker.type);
    if (!unitActions || !unitActions[attackType] || !unitActions[attackType].range) {
        return false;
    }

    const range = unitActions[attackType].range;
    const distance = calculateManhattanDistance(attacker.position, target.position);

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

    // Determine direction and magnitude of move
    let moveX = 0;
    let moveY = 0;

    if (Math.abs(deltaX) > Math.abs(deltaY)) {
        // Prioritize X movement
        moveX = Math.min(maxDistance, Math.abs(deltaX)) * Math.sign(deltaX);
        const remainingMoves = maxDistance - Math.abs(moveX);
        moveY = Math.min(remainingMoves, Math.abs(deltaY)) * Math.sign(deltaY);
    } else {
        // Prioritize Y movement
        moveY = Math.min(maxDistance, Math.abs(deltaY)) * Math.sign(deltaY);
        const remainingMoves = maxDistance - Math.abs(moveY);
        moveX = Math.min(remainingMoves, Math.abs(deltaX)) * Math.sign(deltaX);
    }

    return {
        x: currentUnit.position.x + moveX,
        y: currentUnit.position.y + moveY
    };
}

function isValidPosition(gameState, position) {
    if (!gameState || !gameState.map) return false;

    return position.x >= 0 && position.x < gameState.map.width &&
           position.y >= 0 && position.y < gameState.map.height;
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
<generated>
