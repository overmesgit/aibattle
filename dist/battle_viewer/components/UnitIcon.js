import { html } from 'htm/preact';
import { teamColor, typeSymbol } from '../constants.js';

export const UnitIcon = ({unit}) => {
    return html`
        <div class=${`w-8 h-8 ${teamColor[unit.team]} rounded-full flex items-center justify-center text-white`}>
            ${typeSymbol[unit.type]}
        </div>
    `;
};