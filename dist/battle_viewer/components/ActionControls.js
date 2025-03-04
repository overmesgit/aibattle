import { html } from 'htm/preact';
import { useState, useRef, useEffect } from 'preact/hooks';

export const ActionControls = ({currentActionIndex, setCurrentActionIndex, turns, myTeam}) => {
    const [isAutoPlaying, setIsAutoPlaying] = useState(false);
    const [playbackSpeed, setPlaybackSpeed] = useState(2); // Default medium speed
    const autoPlayRef = useRef(null);

    // Clear autoplay interval when component unmounts
    useEffect(() => {
        return () => {
            if (autoPlayRef.current) {
                clearInterval(autoPlayRef.current);
            }
        };
    }, []);

    // Update autoplay when status or speed changes
    useEffect(() => {
        if (autoPlayRef.current) {
            clearInterval(autoPlayRef.current);
            autoPlayRef.current = null;
        }

        if (isAutoPlaying) {
            // Set speeds: slow (1000ms), medium (500ms), fast (200ms)
            const speedValues = [1000, 500, 200];
            const interval = speedValues[playbackSpeed - 1];

            autoPlayRef.current = setInterval(() => {
                setCurrentActionIndex(current => {
                    if (current >= turns.length - 1) {
                        // Stop when we reach the end
                        setIsAutoPlaying(false);
                        return current;
                    }
                    return current + 1;
                });
            }, interval);
        }

        return () => {
            if (autoPlayRef.current) {
                clearInterval(autoPlayRef.current);
            }
        };
    }, [isAutoPlaying, playbackSpeed, setCurrentActionIndex]);

    const toggleAutoPlay = () => {
        setIsAutoPlaying(!isAutoPlaying);
    };

    return html`
        <div class="bg-white rounded-lg shadow p-4 mb-4">
            <h2 class="text-lg font-bold mb-2 ${myTeam === "teamA" ? 'text-blue-500' : 'text-red-500'}">
                My ${myTeam}
            </h2>
            <h2 class="text-lg font-bold mb-2">
                ${currentActionIndex < turns.length && turns[currentActionIndex].turn ?
                    `Turn ${turns[currentActionIndex].turn}` :
                    "Initial Setup"}
            </h2>
            <h2 class="text-lg font-bold mb-2">Action ${currentActionIndex + 1} of ${turns.length}</h2>
            <input
                type="range"
                min="0"
                max=${turns.length - 1}
                value=${currentActionIndex}
                onChange=${(e) => setCurrentActionIndex(Number(e.target.value))}
                class="w-full mb-2"
            />
            <div class="flex justify-between mb-2">
                <button
                    onClick=${() => setCurrentActionIndex(a => Math.max(0, a - 1))}
                    class="bg-blue-500 text-white px-4 py-2 rounded disabled:bg-gray-400"
                    disabled=${currentActionIndex === 0 || isAutoPlaying}
                >
                    Previous Action
                </button>
                <button
                    onClick=${toggleAutoPlay}
                    class=${isAutoPlaying ? "bg-red-500 text-white px-4 py-2 rounded" : "bg-green-500 text-white px-4 py-2 rounded"}
                    disabled=${currentActionIndex === turns.length - 1}
                >
                    ${isAutoPlaying ? "Pause" : "Play"}
                </button>
                <button
                    onClick=${() => setCurrentActionIndex(a => Math.min(turns.length - 1, a + 1))}
                    class="bg-blue-500 text-white px-4 py-2 rounded disabled:bg-gray-400"
                    disabled=${currentActionIndex === turns.length - 1 || isAutoPlaying}
                >
                    Next Action
                </button>
            </div>
            <div class="flex justify-center space-x-2">
                <span class="text-sm">Speed:</span>
                <button
                    onClick=${() => setPlaybackSpeed(1)}
                    class=${`px-2 py-1 rounded text-sm ${playbackSpeed === 1 ? "bg-blue-500 text-white" : "bg-gray-200"}`}
                >
                    Slow
                </button>
                <button
                    onClick=${() => setPlaybackSpeed(2)}
                    class=${`px-2 py-1 rounded text-sm ${playbackSpeed === 2 ? "bg-blue-500 text-white" : "bg-gray-200"}`}
                >
                    Medium
                </button>
                <button
                    onClick=${() => setPlaybackSpeed(3)}
                    class=${`px-2 py-1 rounded text-sm ${playbackSpeed === 3 ? "bg-blue-500 text-white" : "bg-gray-200"}`}
                >
                    Fast
                </button>
            </div>
        </div>
    `;
};