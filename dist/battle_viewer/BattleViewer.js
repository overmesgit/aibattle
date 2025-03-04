import { html } from 'htm/preact';
import { useState } from 'preact/hooks';
import { css } from '@emotion/css';
import { CELL_SIZE, GRID_SIZE } from './constants.js';
import { BattleGrid } from './components/BattleGrid.js';
import { ActionControls } from './components/ActionControls.js';
import { ActionLog } from './components/ActionLog.js';
import { UnitStatus } from './components/UnitStatus.js';
import { getOffset } from './utils.js';

const gridStyle = css`
    width: ${getOffset(GRID_SIZE, 2)}px;
    height: ${getOffset(GRID_SIZE, 2)}px;`;

export const BattleViewer = ({gameData, myTeam}) => {
    const [currentActionIndex, setCurrentActionIndex] = useState(0);
    const { turns, init_units: initUnits } = gameData;

    return html`
        <div class="flex justify-center m-4 bg-gray-50">
            <div class="bg-gray-50 p-4 ${gridStyle}">
                <${BattleGrid}
                    currentActionIndex=${currentActionIndex}
                    turns=${turns}
                    initUnits=${initUnits}
                />
            </div>
            <div class="p-4 bg-gray-50">
                ${turns.length > 0 && html`
                    <${ActionControls}
                        currentActionIndex=${currentActionIndex}
                        setCurrentActionIndex=${setCurrentActionIndex}
                        turns=${turns}
                        myTeam=${myTeam}
                    />
                `}
                <${ActionLog}
                    currentActionIndex=${currentActionIndex}
                    turns=${turns}
                    initUnits=${initUnits}
                    actionMap=${gameData.actionMap}
                />
            </div>
            <div class="p-4 bg-gray-50">
                <${UnitStatus}
                    currentActionIndex=${currentActionIndex}
                    turns=${turns}
                />
            </div>
        </div>
    `;
};