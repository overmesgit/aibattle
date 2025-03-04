import { CELL_SIZE } from './constants.js';

export const getOffset = (pos, off) => (pos + off) * CELL_SIZE;

export const createActionMap = (gameData) => {
    const actionMap = {};

    if (gameData.unit_action_map) {
        Object.entries(gameData.unit_action_map).forEach(([unitType, actions]) => {
            actionMap[unitType] = {};

            Object.entries(actions).forEach(([actionKey, actionDetails]) => {
                actionMap[unitType][actionKey] = actionDetails.name || actionKey;
            });
        });
    }

    return actionMap;
};