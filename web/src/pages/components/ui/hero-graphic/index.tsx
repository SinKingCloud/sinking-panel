import React from "react";
import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}: any) => ({
    graphic: css`
        width: 100%;
        height: 100%;
        display: block;
        overflow: visible;
        color: ${token.colorPrimary};

        .flow-line {
            fill: none;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        .flow-primary {
            stroke-width: 1.6;
        }

        .flow-secondary {
            stroke: currentColor;
            stroke-width: 1;
            stroke-opacity: .16;
        }

        .flow-dash {
            stroke: currentColor;
            stroke-width: 1;
            stroke-dasharray: 3 10;
            stroke-opacity: .22;
        }

        .flow-signal {
            stroke: currentColor;
            stroke-width: 2.1;
            stroke-dasharray: 15 190;
            stroke-opacity: .48;
            animation: heroSignal 16s linear infinite;
        }

        .node-halo {
            fill: currentColor;
            fill-opacity: .055;
            stroke: currentColor;
            stroke-opacity: .13;
        }

        .node-core {
            fill: currentColor;
            fill-opacity: .52;
        }

        .flow-scene {
            transform-box: fill-box;
            transform-origin: center;
            animation: heroDrift 16s ease-in-out infinite;
        }

        .flow-node-a,
        .flow-node-c {
            animation: heroNode 7s ease-in-out infinite;
        }

        .flow-node-c {
            animation-delay: -3.5s;
        }

        @keyframes heroSignal {
            to {
                stroke-dashoffset: -205;
            }
        }

        @keyframes heroDrift {
            0%, 100% {
                transform: translate3d(0, 0, 0);
            }
            50% {
                transform: translate3d(6px, -3px, 0);
            }
        }

        @keyframes heroNode {
            0%, 100% {
                opacity: .62;
            }
            50% {
                opacity: 1;
            }
        }

        @media (prefers-reduced-motion: reduce) {
            .flow-scene,
            .flow-signal,
            .flow-node-a,
            .flow-node-c {
                animation: none;
            }
        }
    `,
}));

const HeroGraphic = (): React.ReactNode => {
    const {styles} = useStyles();
    const id = `hero-${React.useId().replace(/:/g, "")}`;
    return (
        <svg className={`${styles.graphic} hero-flow-graphic`} viewBox="0 0 480 132"
             aria-hidden="true" focusable="false">
            <defs>
                <linearGradient id={`${id}-flow`} x1="38" y1="90" x2="456" y2="52"
                                gradientUnits="userSpaceOnUse">
                    <stop stopColor="currentColor" stopOpacity="0"/>
                    <stop offset=".24" stopColor="currentColor" stopOpacity=".24"/>
                    <stop offset=".58" stopColor="currentColor" stopOpacity=".7"/>
                    <stop offset="1" stopColor="currentColor" stopOpacity=".04"/>
                </linearGradient>
                <radialGradient id={`${id}-aura`} cx="0" cy="0" r="1"
                                gradientTransform="translate(322 63) rotate(90) scale(60 132)"
                                gradientUnits="userSpaceOnUse">
                    <stop stopColor="currentColor" stopOpacity=".12"/>
                    <stop offset=".55" stopColor="currentColor" stopOpacity=".035"/>
                    <stop offset="1" stopColor="currentColor" stopOpacity="0"/>
                </radialGradient>
            </defs>
            <g className="flow-scene">
                <ellipse className="flow-aura" cx="322" cy="63" rx="132" ry="60"
                         fill={`url(#${id}-aura)`}/>
                <path className="flow-line flow-secondary flow-detail"
                      d="M34 108C116 87 139 29 229 35S346 107 468 52"/>
                <path className="flow-line flow-dash flow-detail"
                      d="M60 30C143 28 174 84 255 86S371 38 454 72"/>
                <path className="flow-line flow-primary" stroke={`url(#${id}-flow)`}
                      d="M26 88C106 84 143 49 215 51S324 94 457 58"/>
                <path className="flow-line flow-signal"
                      d="M26 88C106 84 143 49 215 51S324 94 457 58"/>
                <g className="flow-node flow-node-a" transform="translate(154 57)">
                    <circle className="node-halo" r="13"/>
                    <circle className="node-core" r="3.4"/>
                </g>
                <g className="flow-node flow-node-b flow-secondary" transform="translate(282 78)">
                    <circle className="node-halo" r="10"/>
                    <circle className="node-core" r="2.8"/>
                </g>
                <g className="flow-node flow-node-c" transform="translate(390 67)">
                    <circle className="node-halo" r="14"/>
                    <circle className="node-core" r="3.6"/>
                </g>
            </g>
        </svg>
    );
};

export default React.memo(HeroGraphic);
