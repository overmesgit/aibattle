import {html} from 'htm/preact';
import {useEffect, useRef} from 'preact/hooks';
import {unitTeam} from '../constants.js';

export const ActionLog = ({currentActionIndex, turns, initUnits, actionMap}) => {
    const logRef = useRef(null);

    // Scroll to current action whenever currentActionIndex changes
    useEffect(() => {
        if (logRef.current) {
            // Find the currently selected action element
            const selectedElement = logRef.current.querySelector('.ring-yellow-500');
            if (selectedElement) {
                // Calculate the scroll position to center the element
                const container = logRef.current;
                const containerHeight = container.clientHeight;
                const elementTop = selectedElement.offsetTop;
                const elementHeight = selectedElement.clientHeight;

                // Set scrollTop to center the element in the container
                container.scrollTop = elementTop - 1.5 * containerHeight + (elementHeight / 2);
            }
        }
    }, [currentActionIndex]);

    return html`
        <div class="bg-white rounded-lg shadow p-4 mb-4">
            <h2 class="text-lg font-bold mb-2">Action Log</h2>
            <div ref=${logRef} class="h-80 overflow-y-auto w-96">
                ${turns.map((turn, index) => {
                    if (!turn.unit_action) return null;

                    // Find the unit that performed the action
                    const unit = initUnits.find(u => u.id === turn.unit_id);

                    // Use default background if unit is not found
                    const backGround = unit ?
                            (unit.team === 1 ? "bg-blue-300" : "bg-red-300") :
                            "bg-gray-300";

                    const targetUnit = turn.units_after?.find(u =>
                            u.position.x === turn.unit_action.target?.x &&
                            u.position.y === turn.unit_action.target?.y
                    );

                    // Get the action name from the action map
                    const actionName = actionMap[unit.type]?.[turn.unit_action.action]

                    // Ensure team and unit type are displayed even if not directly available
                    const unitInfo = `${unitTeam[unit.team] || `Team${unit.team}`} ${unit.type || 'Unknown'}: `;

                    return html`
                        <div key=${index} class="m-2">
                            <div class="text-sm p-1 ${backGround} ${index === currentActionIndex ? 'ring-4 ring-yellow-500 ring-offset-0' : ''}">
                                ${turn.turn ? `Turn ${turn.turn}: ` : ''}
                                ${unitInfo}
                                ${actionName}
                                ${turn.unit_action.target ? ` to (${turn.unit_action.target.y}, ${turn.unit_action.target.x})
                                    ${targetUnit ? ` [${targetUnit.type}]` : ''}
                                ` : ''}
                                ${turn.errors ? html`<p class="bg-red-500 p-1">
                        ${turn.errors}</p>` : ''}
                            </div>
                        </div>
                    `;
                })}
            </div>
        </div>
    `;
};