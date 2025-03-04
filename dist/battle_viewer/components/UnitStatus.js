import { html } from 'htm/preact';
import { unitTeam } from '../constants.js';

export const UnitStatus = ({currentActionIndex, turns}) => {
    // Determine the current state of all units at this action index
    const unitMap = new Map();

    // Process all turns up to the current action to get unit states
    for (let i = 0; i <= currentActionIndex && i < turns.length; i++) {
        const turn = turns[i];
        if (turn.units_after) {
            turn.units_after.forEach(unit => {
                unitMap.set(unit.id, unit);
            });
        }
    }

    const units = Array.from(unitMap.values());

    return html`
        <div class="bg-white rounded-lg shadow p-4">
            <h2 class="text-lg font-bold mb-2">Unit Status</h2>
            <div class="space-y-1">
                ${units.map((unit) => html`
                    <div
                        key=${unit.id}
                        class=${`text-sm p-1 ${unit.hp <= 0 ? 'text-gray-400' : ''}
                            ${unit?.team === 1 ? "bg-blue-300" : "bg-red-300"}`}
                    >
                        ${unitTeam[unit?.team]} ${unit.type}: ${Math.max(0, unit.hp)}/${unit.maxHp} HP
                    </div>
                `)}
            </div>
        </div>
    `;
};