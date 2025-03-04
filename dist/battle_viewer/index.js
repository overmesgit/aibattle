import { css } from '@emotion/css';
import { render } from 'preact';
import { html } from 'htm/preact';
import { createActionMap } from './utils.js';
import { BattleViewer } from './BattleViewer.js';

const gameData = window.battleGameData;
console.log(gameData);

// Create action name mapping
gameData.actionMap = createActionMap(gameData);

render(html`
    <${BattleViewer}
        gameData=${gameData}
        myTeam=${window.battleMyTeam}
    />
`, document.getElementById("battle"));