"use client";

const B  = "rgba(37,99,235,";
const A  = "rgba(245,158,11,";
const BG = "rgba(238,242,251,";

const ease = "0.45 0.05 0.55 0.95;0.45 0.05 0.55 0.95";

function CraneJib({
  pivot, dur, begin = "0s", angle = 10, children,
}: {
  pivot: [number, number]; dur: string; begin?: string; angle?: number; children: React.ReactNode;
}) {
  const [cx, cy] = pivot;
  return (
    <g>
      <animateTransform
        attributeName="transform" type="rotate"
        values={`-${angle},${cx},${cy};${angle},${cx},${cy};-${angle},${cx},${cy}`}
        dur={dur} begin={begin} repeatCount="indefinite"
        calcMode="spline" keyTimes="0;0.5;1" keySplines={ease}
      />
      {children}
    </g>
  );
}

function HookRig({
  x, startY, dropY, dur, begin = "0s",
}: {
  x: number; startY: number; dropY: number; dur: string; begin?: string;
}) {
  const cableLen = 44;
  const hookPath = `M${x - 3} ${startY + cableLen} Q${x - 7} ${startY + cableLen} ${x - 7} ${startY + cableLen + 6} Q${x - 7} ${startY + cableLen + 13} ${x} ${startY + cableLen + 13} Q${x + 6} ${startY + cableLen + 10} ${x + 6} ${startY + cableLen + 4} Q${x + 6} ${startY + cableLen} ${x + 3} ${startY + cableLen}`;
  return (
    <g>
      <animateTransform
        attributeName="transform" type="translate"
        values={`0,0;0,${dropY};0,0`}
        dur={dur} begin={begin} repeatCount="indefinite"
        calcMode="spline" keyTimes="0;0.5;1" keySplines={ease}
      />
      <line x1={x} y1={startY} x2={x} y2={startY + cableLen} stroke={`${B}0.48)`} strokeWidth="1"/>
      <path d={hookPath} fill="none" stroke={`${B}0.55)`} strokeWidth="1.3"/>
      <rect x={x - 7} y={startY + cableLen + 13} width="14" height="10" rx="1"
        fill={`${A}0.28)`} stroke={`${A}0.48)`} strokeWidth="0.8"/>
    </g>
  );
}

function Hoist({
  x, bottomY, begin = "0s", dur,
}: {
  x: number; bottomY: number; begin?: string; dur: string;
}) {
  return (
    <>
      <line x1={x} y1={bottomY - 330} x2={x} y2={548} stroke={`${A}0.22)`} strokeWidth="1.1"/>
      <g>
        <animateTransform
          attributeName="transform" type="translate"
          values={`0,0;0,-60;0,-60;0,0`}
          dur={dur} begin={begin} repeatCount="indefinite"
          calcMode="spline" keyTimes="0;0.4;0.6;1"
          keySplines="0.4 0 0.6 1;0 0 1 1;0.4 0 0.6 1"
        />
        <rect x={x - 5} y={bottomY} width="10" height="14" rx="1"
          fill={`${A}0.36)`} stroke={`${A}0.52)`} strokeWidth="0.8"/>
      </g>
    </>
  );
}

function Building({
  x, y, w, h, children,
}: {
  x: number; y: number; w: number; h: number; children?: React.ReactNode;
}) {
  const id = `cl-${x}`;
  return (
    <g>
      <clipPath id={id}><rect x={x} y={y} width={w} height={h}/></clipPath>
      <rect x={x} y={y} width={w} height={h}
        fill={`${B}0.04)`} stroke={`${B}0.20)`} strokeWidth="0.9"/>
      <rect x={x} y={y + 10} width={w} height={h - 10}
        fill="url(#wins)" clipPath={`url(#${id})`}/>
      <line x1={x} y1={y + 5} x2={x + w} y2={y + 5} stroke={`${B}0.26)`} strokeWidth="1.1"/>
      {children}
    </g>
  );
}

export default function Background() {
  return (
    <div className="fixed inset-0 z-0 overflow-hidden pointer-events-none select-none" aria-hidden="true">

      {/* Blueprint dot grid */}
      <div style={{
        position: "absolute", inset: 0,
        backgroundImage: `radial-gradient(${B}0.18) 1.5px, transparent 1.5px)`,
        backgroundSize: "28px 28px",
        animation: "grid-drift 10s linear infinite",
      }}/>

      {/* Orbs */}
      <div className="animate-float1 absolute" style={{
        top: "-8%", left: "-4%", width: "520px", height: "520px",
        borderRadius: "50%", filter: "blur(52px)",
        background: `radial-gradient(circle,${A}0.28) 0%,${A}0.10) 40%,transparent 68%)`,
      }}/>
      <div className="animate-float2 absolute" style={{
        bottom: "-10%", right: "-6%", width: "560px", height: "560px",
        borderRadius: "50%", filter: "blur(56px)",
        background: `radial-gradient(circle,${B}0.20) 0%,${B}0.07) 42%,transparent 68%)`,
      }}/>
      <div className="animate-float3 absolute" style={{
        top: "28%", right: "20%", width: "250px", height: "250px",
        borderRadius: "50%", filter: "blur(30px)",
        background: `radial-gradient(circle,${A}0.14) 0%,transparent 70%)`,
      }}/>

      {/* ── Construction cityscape ─────────────────────────────────── */}
      <svg
        viewBox="0 0 1440 560"
        preserveAspectRatio="xMidYMax meet"
        style={{ position: "absolute", bottom: 0, left: 0, width: "100%", height: "72%" }}
      >
        <defs>
          {/* window pattern */}
          <pattern id="wins" x="0" y="0" width="22" height="20" patternUnits="userSpaceOnUse">
            <rect x="2"  y="2" width="8" height="12" rx="0.5" fill={`${B}0.10)`} stroke={`${B}0.22)`} strokeWidth="0.5"/>
            <rect x="13" y="2" width="8" height="12" rx="0.5" fill={`${B}0.10)`} stroke={`${B}0.22)`} strokeWidth="0.5"/>
          </pattern>
          {/* scaffold pattern */}
          <pattern id="scaf" x="0" y="0" width="20" height="24" patternUnits="userSpaceOnUse">
            <line x1="0"  y1="0"  x2="20" y2="0"  stroke={`${A}0.44)`} strokeWidth="0.9"/>
            <line x1="0"  y1="12" x2="20" y2="12" stroke={`${A}0.44)`} strokeWidth="0.9"/>
            <line x1="0"  y1="24" x2="20" y2="24" stroke={`${A}0.44)`} strokeWidth="0.9"/>
            <line x1="0"  y1="0"  x2="0"  y2="24" stroke={`${A}0.38)`} strokeWidth="0.9"/>
            <line x1="10" y1="0"  x2="10" y2="24" stroke={`${A}0.38)`} strokeWidth="0.9"/>
            <line x1="20" y1="0"  x2="20" y2="24" stroke={`${A}0.38)`} strokeWidth="0.9"/>
            <line x1="0"  y1="0"  x2="10" y2="12" stroke={`${A}0.24)`} strokeWidth="0.7"/>
            <line x1="10" y1="12" x2="20" y2="24" stroke={`${A}0.24)`} strokeWidth="0.7"/>
            <line x1="10" y1="0"  x2="20" y2="12" stroke={`${A}0.24)`} strokeWidth="0.7"/>
            <line x1="0"  y1="12" x2="10" y2="24" stroke={`${A}0.24)`} strokeWidth="0.7"/>
          </pattern>
          {/* fades */}
          <linearGradient id="fl" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%"   stopColor={`${BG}1)`}/>
            <stop offset="14%"  stopColor={`${BG}0)`}/>
          </linearGradient>
          <linearGradient id="fr" x1="0" y1="0" x2="1" y2="0">
            <stop offset="86%"  stopColor={`${BG}0)`}/>
            <stop offset="100%" stopColor={`${BG}1)`}/>
          </linearGradient>
          <linearGradient id="ft" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%"   stopColor={`${BG}0.88)`}/>
            <stop offset="28%"  stopColor={`${BG}0)`}/>
          </linearGradient>
        </defs>

        {/* Ground */}
        <line x1="0" y1="548" x2="1440" y2="548" stroke={`${B}0.22)`} strokeWidth="1.5"/>
        <rect  x="0" y="548" width="1440" height="10" fill={`${B}0.06)`}/>

        {/* ── Completed buildings ──────────────────────────────────── */}
        <Building x={38}   y={355} w={88}  h={193}/>
        <Building x={318}  y={148} w={132} h={400}>
          <line x1="356" y1="148" x2="356" y2="548" stroke={`${B}0.07)`} strokeWidth="0.6"/>
          <line x1="384" y1="148" x2="384" y2="548" stroke={`${B}0.07)`} strokeWidth="0.6"/>
          <line x1="413" y1="148" x2="413" y2="548" stroke={`${B}0.07)`} strokeWidth="0.6"/>
          <rect x="356" y="520" width="40" height="28" fill={`${B}0.07)`} stroke={`${B}0.17)`} strokeWidth="0.7"/>
        </Building>
        <Building x={646}  y={336} w={76}  h={212}/>
        <Building x={952}  y={253} w={100} h={295}>
          <line x1="976"  y1="253" x2="976"  y2="548" stroke={`${B}0.06)`} strokeWidth="0.5"/>
          <line x1="1005" y1="253" x2="1005" y2="548" stroke={`${B}0.06)`} strokeWidth="0.5"/>
          <line x1="1030" y1="253" x2="1030" y2="548" stroke={`${B}0.06)`} strokeWidth="0.5"/>
        </Building>
        <Building x={1248} y={273} w={90}  h={275}/>

        {/* ══════════════════════════════════════════════════════════
            CONSTRUCTION BUILDING 1   mast-top (214, 148)
        ══════════════════════════════════════════════════════════ */}
        <g>
          {/* completed floors */}
          <clipPath id="sc1"><rect x="162" y="355" width="104" height="193"/></clipPath>
          <rect x="162" y="355" width="104" height="193" fill={`${B}0.04)`} stroke={`${B}0.18)`} strokeWidth="0.8"/>
          <rect x="162" y="365" width="104" height="183" fill="url(#wins)" clipPath="url(#sc1)"/>
          {/* scaffold zone */}
          <clipPath id="sf1"><rect x="162" y="228" width="104" height="127"/></clipPath>
          <rect x="162" y="228" width="104" height="127" fill={`${A}0.025)`} stroke={`${A}0.22)`} strokeWidth="0.7" strokeDasharray="5 3"/>
          <rect x="162" y="228" width="104" height="127" fill="url(#scaf)" clipPath="url(#sf1)">
            <animate attributeName="opacity" values="0.7;1;0.7" dur="3.8s" repeatCount="indefinite"/>
          </rect>
          {/* platform */}
          <rect x="156" y="226" width="116" height="4" rx="1" fill={`${A}0.45)`}/>
          {/* mast */}
          <line x1="214" y1="226" x2="214" y2="148" stroke={`${A}0.66)`} strokeWidth="3.5"/>
          <line x1="211" y1="200" x2="217" y2="174" stroke={`${A}0.30)`} strokeWidth="1"/>
          <line x1="217" y1="200" x2="211" y2="174" stroke={`${A}0.30)`} strokeWidth="1"/>
          {/* cab */}
          <rect x="208" y="144" width="13" height="10" rx="1" fill={`${A}0.30)`} stroke={`${A}0.52)`} strokeWidth="0.9"/>
          {/* ANIMATED JIB */}
          <CraneJib pivot={[214, 148]} dur="7s" begin="0s" angle={10}>
            <line x1="214" y1="151" x2="358" y2="151" stroke={`${A}0.60)`} strokeWidth="2.5"/>
            <line x1="214" y1="151" x2="162" y2="158" stroke={`${A}0.55)`} strokeWidth="2"/>
            <line x1="214" y1="148" x2="358" y2="151" stroke={`${A}0.28)`} strokeWidth="1"/>
            <rect x="154" y="157" width="12" height="10" rx="1" fill={`${A}0.40)`}/>
            <rect x="305" y="149" width="11" height="5" rx="0.5" fill={`${A}0.52)`}/>
            <HookRig x={310} startY={154} dropY={34} dur="5s" begin="0.5s"/>
          </CraneJib>
          {/* hoist */}
          <Hoist x={160} bottomY={488} dur="5.5s" begin="1s"/>
        </g>

        {/* ══════════════════════════════════════════════════════════
            CONSTRUCTION BUILDING 2   mast-top (540, 113)
        ══════════════════════════════════════════════════════════ */}
        <g>
          <clipPath id="sc2"><rect x="492" y="328" width="96" height="220"/></clipPath>
          <rect x="492" y="328" width="96" height="220" fill={`${B}0.04)`} stroke={`${B}0.18)`} strokeWidth="0.8"/>
          <rect x="492" y="338" width="96" height="210" fill="url(#wins)" clipPath="url(#sc2)"/>
          <clipPath id="sf2"><rect x="492" y="193" width="96" height="135"/></clipPath>
          <rect x="492" y="193" width="96" height="135" fill={`${A}0.025)`} stroke={`${A}0.20)`} strokeWidth="0.7" strokeDasharray="5 3"/>
          <rect x="492" y="193" width="96" height="135" fill="url(#scaf)" clipPath="url(#sf2)">
            <animate attributeName="opacity" values="1;0.65;1" dur="4.2s" begin="1.3s" repeatCount="indefinite"/>
          </rect>
          <rect x="486" y="191" width="108" height="4" rx="1" fill={`${A}0.44)`}/>
          <line x1="540" y1="191" x2="540" y2="113" stroke={`${A}0.63)`} strokeWidth="3.2"/>
          <line x1="537" y1="164" x2="543" y2="140" stroke={`${A}0.28)`} strokeWidth="1"/>
          <line x1="543" y1="164" x2="537" y2="140" stroke={`${A}0.28)`} strokeWidth="1"/>
          <rect x="534" y="109" width="13" height="10" rx="1" fill={`${A}0.28)`} stroke={`${A}0.48)`} strokeWidth="0.9"/>
          <CraneJib pivot={[540, 113]} dur="9s" begin="1.5s" angle={9}>
            <line x1="540" y1="116" x2="670" y2="116" stroke={`${A}0.58)`} strokeWidth="2.2"/>
            <line x1="540" y1="116" x2="492" y2="122" stroke={`${A}0.53)`} strokeWidth="1.8"/>
            <line x1="540" y1="113" x2="670" y2="116" stroke={`${A}0.26)`} strokeWidth="0.9"/>
            <rect x="485" y="121" width="11" height="9" rx="1" fill={`${A}0.38)`}/>
            <rect x="620" y="114" width="10" height="5" rx="0.5" fill={`${A}0.48)`}/>
            <HookRig x={625} startY={119} dropY={26} dur="6s" begin="2s"/>
          </CraneJib>
          <Hoist x={490} bottomY={475} dur="6s" begin="3s"/>
        </g>

        {/* ══════════════════════════════════════════════════════════
            CONSTRUCTION BUILDING 3  (tallest)  mast-top (832, 28)
        ══════════════════════════════════════════════════════════ */}
        <g>
          <clipPath id="sc3"><rect x="772" y="295" width="120" height="253"/></clipPath>
          <rect x="772" y="295" width="120" height="253" fill={`${B}0.046)`} stroke={`${B}0.22)`} strokeWidth="1"/>
          <rect x="772" y="305" width="120" height="243" fill="url(#wins)" clipPath="url(#sc3)"/>
          <clipPath id="sf3"><rect x="772" y="118" width="120" height="177"/></clipPath>
          <rect x="772" y="118" width="120" height="177" fill={`${A}0.03)`} stroke={`${A}0.23)`} strokeWidth="0.8" strokeDasharray="5 3"/>
          <rect x="772" y="118" width="120" height="177" fill="url(#scaf)" clipPath="url(#sf3)">
            <animate attributeName="opacity" values="0.75;1;0.75" dur="4s" begin="0.8s" repeatCount="indefinite"/>
          </rect>
          <rect x="765" y="116" width="134" height="4.5" rx="1" fill={`${A}0.50)`}/>
          {/* tall mast with braces */}
          <line x1="832" y1="116" x2="832" y2="28"  stroke={`${A}0.72)`} strokeWidth="4.5"/>
          <line x1="828" y1="96"  x2="836" y2="70"  stroke={`${A}0.32)`} strokeWidth="1.2"/>
          <line x1="836" y1="96"  x2="828" y2="70"  stroke={`${A}0.32)`} strokeWidth="1.2"/>
          <line x1="828" y1="68"  x2="836" y2="44"  stroke={`${A}0.32)`} strokeWidth="1.2"/>
          <line x1="836" y1="68"  x2="828" y2="44"  stroke={`${A}0.32)`} strokeWidth="1.2"/>
          {/* cab with window */}
          <rect x="826" y="24"  width="16" height="13" rx="1.5" fill={`${A}0.34)`} stroke={`${A}0.60)`} strokeWidth="1.1"/>
          <rect x="829" y="27"  width="6"  height="5"  rx="0.5" fill={`${B}0.25)`}/>
          {/* ANIMATED JIB — big crane swings wider */}
          <CraneJib pivot={[832, 28]} dur="11s" begin="2.5s" angle={12}>
            <line x1="832" y1="31" x2="1020" y2="31" stroke={`${A}0.67)`} strokeWidth="3"/>
            <line x1="832" y1="31" x2="754"  y2="40" stroke={`${A}0.62)`} strokeWidth="2.5"/>
            <line x1="832" y1="28" x2="1020" y2="31" stroke={`${A}0.33)`} strokeWidth="1.2"/>
            <line x1="832" y1="28" x2="940"  y2="31" stroke={`${A}0.25)`} strokeWidth="1"/>
            <rect x="746" y="39"  width="14" height="12" rx="1" fill={`${A}0.45)`}/>
            <rect x="960" y="29"  width="13" height="6"  rx="0.5" fill={`${A}0.56)`}/>
            <HookRig x={966} startY={35} dropY={44} dur="8s" begin="1s"/>
          </CraneJib>
          <Hoist x={770} bottomY={460} dur="4.5s" begin="0.8s"/>
          <Hoist x={894} bottomY={472} dur="5.8s" begin="2.8s"/>
        </g>

        {/* ══════════════════════════════════════════════════════════
            CONSTRUCTION BUILDING 4   mast-top (1144, 118)
        ══════════════════════════════════════════════════════════ */}
        <g>
          <clipPath id="sc4"><rect x="1098" y="318" width="92" height="230"/></clipPath>
          <rect x="1098" y="318" width="92" height="230" fill={`${B}0.038)`} stroke={`${B}0.18)`} strokeWidth="0.8"/>
          <rect x="1098" y="328" width="92" height="220" fill="url(#wins)" clipPath="url(#sc4)"/>
          <clipPath id="sf4"><rect x="1098" y="198" width="92" height="120"/></clipPath>
          <rect x="1098" y="198" width="92" height="120" fill={`${A}0.023)`} stroke={`${A}0.20)`} strokeWidth="0.7" strokeDasharray="5 3"/>
          <rect x="1098" y="198" width="92" height="120" fill="url(#scaf)" clipPath="url(#sf4)">
            <animate attributeName="opacity" values="0.7;1;0.7" dur="4.5s" begin="2s" repeatCount="indefinite"/>
          </rect>
          <rect x="1092" y="196" width="104" height="4" rx="1" fill={`${A}0.42)`}/>
          <line x1="1144" y1="196" x2="1144" y2="118" stroke={`${A}0.61)`} strokeWidth="3"/>
          <line x1="1141" y1="168" x2="1147" y2="144" stroke={`${A}0.27)`} strokeWidth="1"/>
          <line x1="1147" y1="168" x2="1141" y2="144" stroke={`${A}0.27)`} strokeWidth="1"/>
          <rect x="1138" y="114" width="13" height="10" rx="1" fill={`${A}0.27)`} stroke={`${A}0.47)`} strokeWidth="0.9"/>
          <CraneJib pivot={[1144, 118]} dur="8s" begin="3s" angle={10}>
            <line x1="1144" y1="121" x2="1275" y2="121" stroke={`${A}0.56)`} strokeWidth="2.2"/>
            <line x1="1144" y1="121" x2="1098" y2="127" stroke={`${A}0.51)`} strokeWidth="1.8"/>
            <line x1="1144" y1="118" x2="1275" y2="121" stroke={`${A}0.26)`} strokeWidth="0.9"/>
            <rect x="1091" y="126" width="10" height="9" rx="1" fill={`${A}0.37)`}/>
            <rect x="1222" y="119" width="10" height="5" rx="0.5" fill={`${A}0.47)`}/>
            <HookRig x={1227} startY={124} dropY={30} dur="5s" begin="3.5s"/>
          </CraneJib>
          <Hoist x={1096} bottomY={468} dur="6.5s" begin="1.8s"/>
        </g>

        {/* ── Edge + sky fades ─────────────────────────────────────── */}
        <rect x="0"    y="0" width="150"  height="560" fill="url(#fl)"/>
        <rect x="1290" y="0" width="150"  height="560" fill="url(#fr)"/>
        <rect x="0"    y="0" width="1440" height="210" fill="url(#ft)"/>
      </svg>

      {/* Ground fade */}
      <div style={{
        position: "absolute", bottom: 0, left: 0, right: 0, height: "70px",
        background: `linear-gradient(to top,${BG}0.75),transparent)`,
      }}/>
    </div>
  );
}
