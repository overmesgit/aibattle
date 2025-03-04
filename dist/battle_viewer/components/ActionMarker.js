import { html } from 'htm/preact';
import { css } from '@emotion/css';
import { teamColor, typeSymbol } from '../constants.js';
import { getOffset } from '../utils.js';

export const ActionMarker = ({action, unit}) => {
    if (!action.target || action.action === "hold" || action.action === "move") return null;

    const color = "bg-grey-300";
    const symbol = action.action.includes('attack') ? '⚔️' :
        action.action.includes('skill') ? '✨' : '⚡';
    const markerStyle = css`
        left: ${getOffset(action.target.x, 1)}px;
        top: ${getOffset(action.target.y, 1)}px;`;

    return html`
        <div class="absolute z-10 opacity-90 ${markerStyle}">
            <div class=${`w-4 h-4 ${color} rounded-full flex items-center justify-center text-white`}>
                ${symbol}
            </div>
            <div class=${`w-4 h-4 ${teamColor[unit.team]} rounded-full flex items-center justify-center text-white`}>
                ${typeSymbol[unit.type]}
            </div>
        </div>
    `;
};