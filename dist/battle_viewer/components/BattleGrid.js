import { html } from 'htm/preact';
import { css } from '@emotion/css';
import { GridCoordinates, GridLines } from './GridComponents.js';
import { ActionMarker } from './ActionMarker.js';
import { UnitIcon } from './UnitIcon.js';
import { getOffset } from '../utils.js';

export const BattleGrid = ({currentActionIndex, turns, initUnits}) => {
    // Track units and their positions
    let units = [];

    // Create a map of unit IDs to the most recent unit state
    const unitMap = new Map();

    // First, initialize with all initial units
    if (initUnits && initUnits.length > 0) {
        initUnits.forEach(unit => {
            unitMap.set(unit.id, unit);
        });
    }

    // Process all turns up to the current action index to get final unit positions
    for (let i = 0; i <= currentActionIndex && i < turns.length; i++) {
        const turn = turns[i];

        // If units_after is available, update the unit state
        if (turn.units_after) {
            turn.units_after.forEach(unit => {
                unitMap.set(unit.id, unit);
            });
        }
    }

    // Convert the map to an array for rendering
    units = Array.from(unitMap.values());

    // Get the current turn for display
    const currentAction = currentActionIndex < turns.length ? turns[currentActionIndex] : null;

    // Get the current unit ID if available
    const currentUnitId = currentAction ? currentAction.unit_id : null;

    const unitStyle = (unit) => css`
        left: ${getOffset(unit.position.x, 1)}px;
        top: ${getOffset(unit.position.y, 1)}px;`;

    return html`
        <div>
            <div class="relative">
                <${GridCoordinates}/>
                <${GridLines}/>
                ${currentAction && currentAction.unit_action && html`
                    <${ActionMarker}
                            action=${currentAction.unit_action}
                            unit=${unitMap.get(currentAction.unit_id)}/>
                `}
                ${units.map((unit) =>
                    unit.hp > 0 && html`
                        <div key=${unit.id} class="absolute ${unitStyle(unit)}">
                            <div class=${unit.id === currentUnitId ? 'ring-4 ring-yellow-500 ring-offset-0 rounded-full' : ''}>
                                <${UnitIcon} unit=${unit}/>
                            </div>
                        </div>
                    `
                )}
            </div>
        </div>
    `;
};