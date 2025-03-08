import { html } from 'htm/preact';
import { useState } from 'preact/hooks';
import { css } from '@emotion/css';
import { CELL_SIZE, GRID_SIZE } from './constants.js';
import { BattleGrid } from './components/BattleGrid.js';
import { ActionControls } from './components/ActionControls.js';
import { ActionLog } from './components/ActionLog.js';
import { UnitStatus } from './components/UnitStatus.js';

const styles = {
    container: css`
        display: flex;
        flex-direction: row;
        gap: 1rem;
        margin: 1rem;
        background: rgb(249 250 251);
        max-width: 100%;
        
        @media (max-width: 1600px) {
            flex-direction: column;
            align-items: center;
        }
    `,
    
    gridWrapper: css`
        padding: 1rem;
        background: rgb(249 250 251);
        min-width: 700px;
        min-height: 700px;
    `,
    
    grid: css`
        width: 100%;
        aspect-ratio: 1;
    `,
    
    sidePanel: css`
        padding: 1rem;
        background: rgb(249 250 251);
        
        @media (max-width: 1600px) {
            width: 100%;
            max-width: 700px;
            padding: 0.5rem;
        }
    `
};

export const BattleViewer = ({gameData, myTeam}) => {
    const [currentActionIndex, setCurrentActionIndex] = useState(0);
    const { turns, init_units: initUnits } = gameData;

    return html`
        <div class=${styles.container}>
            <div class=${styles.gridWrapper}>
                <div class=${styles.grid}>
                    <${BattleGrid}
                        currentActionIndex=${currentActionIndex}
                        turns=${turns}
                        initUnits=${initUnits}
                    />
                </div>
            </div>
            <div class=${styles.sidePanel}>
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
            <div class=${styles.sidePanel}>
                <${UnitStatus}
                    currentActionIndex=${currentActionIndex}
                    turns=${turns}
                />
            </div>
        </div>
    `;
};