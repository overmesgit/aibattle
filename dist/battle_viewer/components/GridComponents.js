import { html } from 'htm/preact';
import { css } from '@emotion/css';
import { CELL_SIZE, GRID_SIZE } from '../constants.js';
import { getOffset } from '../utils.js';

const gridStyle = css`
    width: ${getOffset(GRID_SIZE, 2)}px;
    height: ${getOffset(GRID_SIZE, 2)}px;`;

export const GridCoordinates = () => html`
    ${Array.from({length: GRID_SIZE}).map((_, i) => html`
        <div key=${`numbers-${i}`}>
            <div class=${`absolute bg-gray-800 text-white flex items-center justify-center w-8 h-8
            ${css` left: ${getOffset(i, 1)}px; top: 0; `}`}>
                ${i}
            </div>
            <div class=${`absolute bg-gray-800 text-white flex items-center justify-center w-8 h-8
             ${css` left: 0; top: ${getOffset(i, 1)}px; `}`}>
                ${i}
            </div>
        </div>
    `)}
`;

const linesLength = (GRID_SIZE + 1) * CELL_SIZE;

export const GridLines = () => html`
    <div class="absolute" style=${gridStyle}>
        ${Array.from({length: GRID_SIZE + 1}).map((_, i) => html`
            <div key=${`grid-${i}`}>
                <div class=${`absolute border-t border-gray-300
                        ${css` left: 0; top: ${getOffset(i, 1)}px; width: ${linesLength}px; `}`}></div>
                <div class=${`absolute border-l border-gray-300
                        ${css` left: ${getOffset(i, 1)}px; top: 0; height: ${linesLength}px; `}`}></div>
            </div>
        `)}
    </div>
`;