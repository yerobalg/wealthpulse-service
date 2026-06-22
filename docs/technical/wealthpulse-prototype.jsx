import { useState } from "react";

/* ─── Constants ─── */
const CATEGORIES = [
  { key: "crypto", label: "Crypto", color: "#F7931A", icon: "₿" },
  { key: "idx", label: "IDX Stocks", color: "#E63946", icon: "🇮🇩" },
  { key: "us", label: "US Stocks", color: "#457B9D", icon: "🇺🇸" },
  { key: "gold", label: "Gold", color: "#D4A843", icon: "✦" },
  { key: "bonds", label: "IDN Bonds", color: "#2A9D8F", icon: "📜" },
  { key: "cash", label: "Cash", color: "#8B8FA3", icon: "💵" },
];

const TARGET_ALLOC = { crypto: 20, idx: 30, us: 20, gold: 15, bonds: 15, cash: 0 };

const SAMPLE_TRANSACTIONS = [
  { id: 1, date: "2025-08-15", type: "buy", category: "crypto", ticker: "BTC", name: "Bitcoin", qty: 0.15, price: 920000000, notes: "" },
  { id: 2, date: "2025-09-02", type: "buy", category: "crypto", ticker: "ETH", name: "Ethereum", qty: 2.5, price: 28400000, notes: "" },
  { id: 3, date: "2025-09-20", type: "buy", category: "crypto", ticker: "SOL", name: "Solana", qty: 120, price: 1460000, notes: "" },
  { id: 4, date: "2025-07-10", type: "buy", category: "idx", ticker: "BBCA", name: "Bank Central Asia", qty: 5000, price: 9200, notes: "" },
  { id: 5, date: "2025-07-22", type: "buy", category: "idx", ticker: "TLKM", name: "Telkom Indonesia", qty: 20000, price: 3800, notes: "" },
  { id: 6, date: "2025-08-05", type: "buy", category: "idx", ticker: "BBRI", name: "Bank Rakyat Ind.", qty: 15000, price: 5100, notes: "" },
  { id: 7, date: "2025-10-01", type: "buy", category: "idx", ticker: "ASII", name: "Astra International", qty: 10000, price: 6200, notes: "" },
  { id: 8, date: "2025-06-15", type: "buy", category: "us", ticker: "AAPL", name: "Apple Inc.", qty: 25, price: 178, notes: "" },
  { id: 9, date: "2025-07-01", type: "buy", category: "us", ticker: "NVDA", name: "NVIDIA Corp.", qty: 15, price: 620, notes: "" },
  { id: 10, date: "2025-08-20", type: "buy", category: "us", ticker: "VOO", name: "Vanguard S&P 500", qty: 5, price: 445, notes: "" },
  { id: 11, date: "2025-05-01", type: "buy", category: "gold", ticker: "GOLD", name: "Antam Gold", qty: 200, price: 1100000, notes: "200 gram" },
  { id: 12, date: "2025-04-01", type: "buy", category: "bonds", ticker: "ORI024", name: "ORI024", qty: 1, price: 150000000, notes: "Annual return: 6.25%" },
  { id: 13, date: "2025-04-15", type: "buy", category: "bonds", ticker: "SR020", name: "SR020", qty: 1, price: 120000000, notes: "Annual return: 6.4%" },
  { id: 14, date: "2026-01-10", type: "buy", category: "cash", ticker: "IDR", name: "Indonesian Rupiah", qty: 50000000, price: 1, notes: "Emergency fund" },
  { id: 15, date: "2026-01-10", type: "buy", category: "cash", ticker: "USD", name: "US Dollar", qty: 2000, price: 1, notes: "USD savings" },
];

const CURRENT_PRICES = {
  BTC: 1020000000, ETH: 31200000, SOL: 1462000,
  BBCA: 9850, TLKM: 4120, BBRI: 5400, ASII: 6050,
  AAPL: 2901250, NVDA: 12544000, VOO: 7176000,
  GOLD: 1385625, ORI024: 152250000, SR020: 124875000,
  IDR: 1, USD: 16250,
};
const USD_IDR = 16250;

const fmtRp = (v) => { const a = Math.abs(v), s = v < 0 ? "-" : ""; if (a >= 1e12) return `${s}Rp ${(a/1e12).toFixed(2)}T`; if (a >= 1e9) return `${s}Rp ${(a/1e9).toFixed(2)}B`; if (a >= 1e6) return `${s}Rp ${(a/1e6).toFixed(1)}M`; if (a >= 1e3) return `${s}Rp ${(a/1e3).toFixed(0)}K`; return `${s}Rp ${a.toLocaleString("id-ID")}`; };
const fmtUsd = (v) => { const a = Math.abs(v), s = v < 0 ? "-" : ""; if (a >= 1e9) return `${s}$${(a/1e9).toFixed(2)}B`; if (a >= 1e6) return `${s}$${(a/1e6).toFixed(2)}M`; if (a >= 1e3) return `${s}$${(a/1e3).toFixed(0)}K`; return `${s}$${a.toFixed(0)}`; };
const toIDR = (v, c) => c === "us" ? v * USD_IDR : v;
const fmtVal = (v, c) => c === "USD" ? fmtUsd(v / USD_IDR) : fmtRp(v);

const computeHoldings = (txns) => {
  const m = {};
  txns.forEach(t => {
    if (!m[t.ticker]) m[t.ticker] = { ticker: t.ticker, name: t.name, category: t.category, totalQty: 0, totalCost: 0 };
    const p = toIDR(t.price, t.category);
    if (t.type === "buy") { m[t.ticker].totalQty += t.qty; m[t.ticker].totalCost += t.qty * p; } else m[t.ticker].totalQty -= t.qty;
  });
  return Object.values(m).filter(h => h.totalQty > 0).map(h => {
    const cp = toIDR(CURRENT_PRICES[h.ticker] || 0, h.category), cv = h.totalQty * cp;
    const pnl = cv - h.totalCost, pnlPct = h.totalCost > 0 ? (pnl / h.totalCost) * 100 : 0;
    return { ...h, avgPrice: h.totalCost / h.totalQty, curPrice: cp, currentValue: cv, pnl, pnlPct };
  });
};

const Spark = ({ up }) => <svg width="50" height="22" viewBox="0 0 54 22"><polyline points={up ? "0,18 6,16 12,13 18,15 24,10 30,8 36,11 42,6 48,4 54,2" : "0,3 6,5 12,9 18,7 24,13 30,15 36,12 42,16 48,18 54,19"} fill="none" stroke={up ? "#2A9D8F" : "#E63946"} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" /></svg>;

const Donut = ({ segments, size = 150, centerText, centerSub }) => {
  const cx = size/2, cy = size/2, r = size*0.35, sw = size*0.12, circ = 2*Math.PI*r; let cum = 0;
  return <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
    {segments.filter(s=>s.pct>0).map((s,i) => { const d=(s.pct/100)*circ, o=-(cum/100)*circ; cum+=s.pct; return <circle key={i} cx={cx} cy={cy} r={r} fill="none" stroke={s.color} strokeWidth={sw} strokeDasharray={`${d} ${circ-d}`} strokeDashoffset={o} transform={`rotate(-90 ${cx} ${cy})`} style={{transition:"all 0.4s"}} />; })}
    <text x={cx} y={cy-5} textAnchor="middle" fill="#EAEDF6" fontSize="12" fontWeight="700" fontFamily="'JetBrains Mono',monospace">{centerText}</text>
    <text x={cx} y={cy+10} textAnchor="middle" fill="#6B7084" fontSize="9">{centerSub}</text>
  </svg>;
};

const SectionHead = ({ cat, value, pnlPct, count, cur }) => (
  <div style={{ display:"flex", alignItems:"center", justifyContent:"space-between", padding:"9px 14px", background:`${cat.color}08`, borderBottom:`2px solid ${cat.color}25`, borderRadius:"10px 10px 0 0" }}>
    <div style={{ display:"flex", alignItems:"center", gap:7 }}><span style={{fontSize:13}}>{cat.icon}</span><span style={{fontSize:13,fontWeight:700,color:"#EAEDF6"}}>{cat.label}</span><span style={{fontSize:9,color:"#3A3E50",fontFamily:"monospace"}}>{count} assets</span></div>
    <div style={{ display:"flex", alignItems:"center", gap:10 }}>
      <span style={{fontSize:12,fontWeight:600,color:"#EAEDF6",fontFamily:"'JetBrains Mono',monospace"}}>{fmtVal(value,cur)}</span>
      <span style={{fontSize:9,fontWeight:600,padding:"2px 6px",borderRadius:5,background:pnlPct>=0?"#2A9D8F12":"#E6394612",color:pnlPct>=0?"#2A9D8F":"#E63946",fontFamily:"'JetBrains Mono',monospace"}}>{pnlPct>=0?"▲":"▼"} {Math.abs(pnlPct).toFixed(1)}%</span>
    </div>
  </div>
);

export default function PortfolioIQ() {
  const [page, setPage] = useState("dashboard");
  const [cur, setCur] = useState("IDR");
  const [transactions, setTransactions] = useState(SAMPLE_TRANSACTIONS);
  const [showAddTx, setShowAddTx] = useState(false);
  const [txForm, setTxForm] = useState({ type:"buy", category:"crypto", ticker:"", name:"", qty:"", price:"", notes:"", annualReturn:"" });
  const [txPage, setTxPage] = useState(1);
  const [txPerPage, setTxPerPage] = useState(10);
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const holdings = computeHoldings(transactions);
  const totalValue = holdings.reduce((s,h)=>s+h.currentValue,0);
  const totalCost = holdings.reduce((s,h)=>s+h.totalCost,0);
  const totalPnl = totalValue - totalCost;
  const totalPnlPct = totalCost > 0 ? (totalPnl/totalCost)*100 : 0;

  const allocByCategory = CATEGORIES.map(cat => {
    const ch = holdings.filter(h=>h.category===cat.key), cv = ch.reduce((s,h)=>s+h.currentValue,0), cc = ch.reduce((s,h)=>s+h.totalCost,0);
    return { ...cat, value:cv, pnl:cv-cc, pnlPct:cc>0?((cv-cc)/cc)*100:0, pct:totalValue>0?(cv/totalValue)*100:0, target:TARGET_ALLOC[cat.key]||0, count:ch.length };
  }).filter(a=>a.value>0);

  const alerts = [
    { type:"up", asset:"BTC", cond:"≥ Rp 1.05B", active:true },
    { type:"down", asset:"BBCA", cond:"≤ Rp 9,000", active:true },
    { type:"pct", asset:"Portfolio", cond:"Day Δ ≤ -3%", active:true },
    { type:"up", asset:"NVDA", cond:"≥ $900", active:true },
    { type:"down", asset:"SOL", cond:"≤ Rp 1.2M", active:false },
  ];

  const filteredTx = transactions.filter(t=>!dateFrom||t.date>=dateFrom).filter(t=>!dateTo||t.date<=dateTo).sort((a,b)=>b.date.localeCompare(a.date));
  const txTotalPages = Math.ceil(filteredTx.length/txPerPage);
  const txSlice = filteredTx.slice((txPage-1)*txPerPage, txPage*txPerPage);

  const resetForm = () => setTxForm({ type:"buy", category:"crypto", ticker:"", name:"", qty:"", price:"", notes:"", annualReturn:"" });
  const handleAddTx = () => {
    if (txForm.category==="cash" && (!txForm.ticker||!txForm.qty)) return;
    if (txForm.category==="bonds" && (!txForm.name||!txForm.qty||!txForm.price)) return;
    if (!["cash","bonds"].includes(txForm.category) && (!txForm.ticker||!txForm.qty||!txForm.price)) return;
    setTransactions([...transactions, {
      id:Date.now(), date:new Date().toISOString().slice(0,10), type:txForm.type, category:txForm.category,
      ticker: txForm.category==="bonds" ? txForm.name.toUpperCase() : txForm.ticker.toUpperCase(),
      name: txForm.category==="bonds" ? txForm.name : (txForm.name||txForm.ticker.toUpperCase()),
      qty:parseFloat(txForm.qty), price:txForm.category==="cash"?1:parseFloat(txForm.price),
      notes: txForm.category==="bonds"&&txForm.annualReturn ? `Annual return: ${txForm.annualReturn}%` : txForm.notes,
    }]);
    resetForm(); setShowAddTx(false);
  };

  const I = { width:"100%", padding:"8px 10px", borderRadius:7, border:"1px solid #1E2130", background:"#0A0C12", color:"#EAEDF6", fontSize:12, fontFamily:"'DM Sans',sans-serif", outline:"none" };
  const L = { fontSize:10, color:"#6B7084", marginBottom:4, display:"block" };
  const B = (bg,fg) => ({ padding:"7px 14px", borderRadius:7, border:"none", background:bg, color:fg, fontSize:11, fontWeight:600, cursor:"pointer", fontFamily:"inherit" });
  const canSubmit = txForm.category==="cash"?(txForm.ticker&&txForm.qty):txForm.category==="bonds"?(txForm.name&&txForm.qty&&txForm.price):(txForm.ticker&&txForm.qty&&txForm.price);

  return (
    <div style={{minHeight:"100vh",background:"#0A0C12",color:"#C8CCD8",fontFamily:"'DM Sans',system-ui,sans-serif"}}>
      <style>{`@import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap');*{box-sizing:border-box;margin:0;padding:0}input:focus,select:focus{border-color:#2A9D8F!important}::-webkit-scrollbar{width:5px}::-webkit-scrollbar-thumb{background:#1E2130;border-radius:3px}`}</style>

      {/* Top Bar */}
      <div style={{padding:"11px 20px",borderBottom:"1px solid #13151D",display:"flex",justifyContent:"space-between",alignItems:"center",position:"sticky",top:0,background:"#0A0C12ee",backdropFilter:"blur(10px)",zIndex:20}}>
        <div style={{display:"flex",alignItems:"center",gap:8}}>
          <div style={{width:28,height:28,borderRadius:7,background:"linear-gradient(135deg,#F7931A,#D4A843)",display:"flex",alignItems:"center",justifyContent:"center",fontSize:13,fontWeight:800,color:"#0A0C12"}}>P</div>
          <span style={{fontSize:14,fontWeight:700,color:"#EAEDF6",letterSpacing:"-0.02em"}}>PortfolioIQ</span>
        </div>
        <div style={{display:"flex",alignItems:"center",gap:6}}>
          {["dashboard","transactions"].map(p=>(
            <button key={p} onClick={()=>{setPage(p);setTxPage(1)}} style={{padding:"5px 12px",borderRadius:6,border:"1px solid",borderColor:page===p?"#D4A84340":"#1A1D2A",background:page===p?"#D4A84310":"transparent",color:page===p?"#D4A843":"#4A4E60",fontSize:11,fontWeight:600,cursor:"pointer",fontFamily:"inherit"}}>{p==="dashboard"?"📊 Dashboard":"📋 Transactions"}</button>
          ))}
          <div style={{width:1,height:20,background:"#1A1D2A",margin:"0 4px"}} />
          <div style={{display:"flex",borderRadius:6,border:"1px solid #1A1D2A",overflow:"hidden"}}>
            {["IDR","USD"].map(c=>(<button key={c} onClick={()=>setCur(c)} style={{padding:"4px 10px",border:"none",background:cur===c?"#2A9D8F15":"transparent",color:cur===c?"#2A9D8F":"#3A3E50",fontSize:10,fontWeight:600,cursor:"pointer",fontFamily:"'JetBrains Mono',monospace"}}>{c}</button>))}
          </div>
          <div style={{padding:"4px 8px",borderRadius:6,background:"#229ED910",fontSize:9,color:"#229ED9",fontWeight:600}}>✈ Telegram</div>
        </div>
      </div>

      <div style={{padding:"16px 20px",maxWidth:1100,margin:"0 auto"}}>

        {/* ═══ DASHBOARD ═══ */}
        {page==="dashboard"&&(<>
          <div style={{display:"grid",gridTemplateColumns:"repeat(5,1fr)",gap:8,marginBottom:14}}>
            {[{label:"Total Value",val:fmtVal(totalValue,cur),color:"#EAEDF6"},{label:"Total Cost",val:fmtVal(totalCost,cur),color:"#8B8FA3"},{label:"Total P&L",val:`${totalPnl>=0?"+":""}${fmtVal(totalPnl,cur)}`,color:totalPnl>=0?"#2A9D8F":"#E63946",sub:`${totalPnlPct>=0?"+":""}${totalPnlPct.toFixed(2)}%`},{label:"Today",val:"+Rp 12.4M",color:"#2A9D8F",sub:"+0.67%"},{label:"YoY Progress",val:`${totalPnlPct.toFixed(1)}%`,color:"#D4A843",sub:"of 15% target"}].map((c,i)=>(
              <div key={i} style={{padding:"12px 14px",borderRadius:10,background:"#0F1119",border:"1px solid #16182230"}}>
                <div style={{fontSize:9,color:"#3A3E50",textTransform:"uppercase",letterSpacing:"0.06em",fontWeight:600,marginBottom:5}}>{c.label}</div>
                <div style={{fontSize:15,fontWeight:700,color:c.color,fontFamily:"'JetBrains Mono',monospace",letterSpacing:"-0.02em"}}>{c.val}</div>
                {c.sub&&<div style={{fontSize:10,color:c.color,opacity:0.6,fontFamily:"'JetBrains Mono',monospace",marginTop:2}}>{c.sub}</div>}
              </div>
            ))}
          </div>

          <div style={{display:"grid",gridTemplateColumns:"170px 1fr 210px",gap:10,marginBottom:14}}>
            <div style={{padding:"12px",borderRadius:12,background:"#0F1119",border:"1px solid #16182230",display:"flex",flexDirection:"column",alignItems:"center",justifyContent:"center"}}>
              <Donut segments={allocByCategory} size={140} centerText={fmtVal(totalValue,cur)} centerSub="Total Value" />
              <div style={{width:"100%",marginTop:10}}>{allocByCategory.map(a=>(<div key={a.key} style={{display:"flex",alignItems:"center",gap:6,marginBottom:5}}><div style={{width:6,height:6,borderRadius:2,background:a.color}} /><span style={{fontSize:10,color:"#8B8FA3",flex:1}}>{a.label}</span><span style={{fontSize:10,color:"#EAEDF6",fontWeight:600,fontFamily:"'JetBrains Mono',monospace"}}>{a.pct.toFixed(0)}%</span></div>))}</div>
            </div>
            <div style={{padding:"14px 16px",borderRadius:12,background:"#0F1119",border:"1px solid #16182230"}}>
              <div style={{fontSize:9,color:"#3A3E50",textTransform:"uppercase",letterSpacing:"0.06em",fontWeight:600,marginBottom:10}}>Actual vs Target Allocation</div>
              {allocByCategory.filter(a=>a.target>0).map(a=>{const d=a.pct-a.target;return(
                <div key={a.key} style={{marginBottom:10}}>
                  <div style={{display:"flex",justifyContent:"space-between",alignItems:"center",marginBottom:3}}>
                    <div style={{display:"flex",alignItems:"center",gap:6}}><div style={{width:6,height:6,borderRadius:2,background:a.color}} /><span style={{fontSize:11,color:"#C8CCD8"}}>{a.label}</span></div>
                    <div style={{display:"flex",alignItems:"center",gap:6}}><span style={{fontSize:10,color:"#6B7084",fontFamily:"'JetBrains Mono',monospace"}}>{a.pct.toFixed(0)}% / {a.target}%</span><span style={{fontSize:9,fontWeight:600,padding:"1px 5px",borderRadius:3,background:Math.abs(d)<=2?"#2A9D8F10":"#E6394610",color:Math.abs(d)<=2?"#2A9D8F":d>0?"#F7931A":"#E63946",fontFamily:"'JetBrains Mono',monospace"}}>{d>=0?"+":""}{d.toFixed(0)}%</span></div>
                  </div>
                  <div style={{height:5,background:"#14161F",borderRadius:3,position:"relative"}}><div style={{width:`${Math.min(a.pct,100)}%`,height:"100%",background:a.color,borderRadius:3,transition:"width 0.4s"}} /><div style={{position:"absolute",top:-2,left:`${a.target}%`,width:1.5,height:9,background:"#EAEDF640",borderRadius:1}} /></div>
                </div>
              );})}
              {allocByCategory.some(a=>a.target>0&&Math.abs(a.pct-a.target)>2)&&(<div style={{marginTop:8,padding:"8px 10px",borderRadius:7,background:"#D4A84308",border:"1px solid #D4A84312",fontSize:10,color:"#8B8FA3",lineHeight:1.5}}><span style={{color:"#D4A843",fontWeight:600}}>💡 Rebalance: </span>{allocByCategory.filter(a=>a.target>0&&a.pct-a.target<-2).map(a=>`${a.label} is ${(a.target-a.pct).toFixed(0)}% below target`).join(". ")}</div>)}
            </div>
            <div style={{padding:"14px 16px",borderRadius:12,background:"#0F1119",border:"1px solid #16182230"}}>
              <div style={{display:"flex",justifyContent:"space-between",alignItems:"center",marginBottom:8}}><span style={{fontSize:9,color:"#3A3E50",textTransform:"uppercase",letterSpacing:"0.06em",fontWeight:600}}>🔔 Alerts</span><span style={{fontSize:9,color:"#229ED9",fontWeight:600,cursor:"pointer"}}>+ Add</span></div>
              {alerts.map((a,i)=>(<div key={i} style={{display:"flex",alignItems:"center",gap:6,padding:"6px 0",borderBottom:i<alerts.length-1?"1px solid #14161F":"none",opacity:a.active?1:0.35}}><span style={{fontSize:10}}>{a.type==="up"?"📈":a.type==="down"?"📉":"⚡"}</span><span style={{fontSize:11,fontWeight:600,color:"#EAEDF6",minWidth:36}}>{a.asset}</span><span style={{fontSize:10,color:"#4A4E60",fontFamily:"'JetBrains Mono',monospace",flex:1}}>{a.cond}</span><div style={{width:5,height:5,borderRadius:3,background:a.active?"#2A9D8F":"#2A2D3A",boxShadow:a.active?"0 0 5px #2A9D8F30":"none"}} /></div>))}
            </div>
          </div>

          <div style={{padding:"9px 16px",borderRadius:8,marginBottom:14,background:"linear-gradient(90deg,#D4A84306,#2A9D8F06)",border:"1px solid #D4A84312",display:"flex",alignItems:"center",gap:12}}>
            <span style={{fontSize:10,color:"#D4A843",fontWeight:600,whiteSpace:"nowrap"}}>YoY Target: 15%</span>
            <div style={{flex:1,height:5,background:"#14161F",borderRadius:3}}><div style={{width:`${Math.min((totalPnlPct/15)*100,100)}%`,height:"100%",borderRadius:3,background:"linear-gradient(90deg,#D4A843,#2A9D8F)"}} /></div>
            <span style={{fontSize:11,fontWeight:700,color:"#2A9D8F",fontFamily:"'JetBrains Mono',monospace"}}>{totalPnlPct.toFixed(1)}%</span>
            <span style={{fontSize:9,color:"#3A3E50"}}>({((totalPnlPct/15)*100).toFixed(0)}%)</span>
          </div>

          <div style={{display:"flex",justifyContent:"flex-end",marginBottom:10}}><button onClick={()=>setShowAddTx(true)} style={{...B("#2A9D8F","#0A0C12")}}>+ Add Transaction</button></div>

          {CATEGORIES.filter(cat=>holdings.some(h=>h.category===cat.key)).map(cat=>{
            const ch=holdings.filter(h=>h.category===cat.key),cv=ch.reduce((s,h)=>s+h.currentValue,0),cc=ch.reduce((s,h)=>s+h.totalCost,0);
            return(<div key={cat.key} style={{marginBottom:8,borderRadius:10,overflow:"hidden",background:"#0F1119",border:"1px solid #16182230"}}>
              <SectionHead cat={cat} value={cv} pnlPct={cc>0?((cv-cc)/cc)*100:0} count={ch.length} cur={cur} />
              <div style={{display:"grid",gridTemplateColumns:"minmax(110px,2fr) minmax(50px,1fr) minmax(70px,1.2fr) minmax(70px,1.2fr) minmax(80px,1.3fr) 50px",padding:"5px 14px",fontSize:9,color:"#2A2D3A",textTransform:"uppercase",letterSpacing:"0.06em",fontWeight:600,borderBottom:"1px solid #14161F"}}><span>Asset</span><span style={{textAlign:"right"}}>Qty</span><span style={{textAlign:"right"}}>Avg Price</span><span style={{textAlign:"right"}}>Current</span><span style={{textAlign:"right"}}>P&L</span><span /></div>
              {ch.map((h,i)=>(<div key={i} style={{display:"grid",gridTemplateColumns:"minmax(110px,2fr) minmax(50px,1fr) minmax(70px,1.2fr) minmax(70px,1.2fr) minmax(80px,1.3fr) 50px",padding:"9px 14px",alignItems:"center",borderBottom:i<ch.length-1?"1px solid #10121A":"none"}}>
                <div><span style={{fontSize:12,fontWeight:600,color:"#EAEDF6"}}>{h.ticker}</span><span style={{fontSize:9,color:"#3A3E50",marginLeft:5}}>{h.name}</span></div>
                <div style={{textAlign:"right",fontSize:11,color:"#6B7084",fontFamily:"'JetBrains Mono',monospace"}}>{h.category==="cash"?fmtVal(h.totalQty,h.ticker==="USD"?"USD":"IDR"):h.totalQty.toLocaleString()}</div>
                <div style={{textAlign:"right",fontSize:10,color:"#3A3E50",fontFamily:"'JetBrains Mono',monospace"}}>{h.category==="cash"?"—":fmtVal(h.avgPrice,cur)}</div>
                <div style={{textAlign:"right",fontSize:10,color:"#8B8FA3",fontFamily:"'JetBrains Mono',monospace"}}>{h.category==="cash"?"—":fmtVal(h.curPrice,cur)}</div>
                <div style={{textAlign:"right"}}>{h.category==="cash"?<span style={{fontSize:10,color:"#3A3E50"}}>—</span>:<><div style={{fontSize:11,fontWeight:600,color:h.pnlPct>=0?"#2A9D8F":"#E63946",fontFamily:"'JetBrains Mono',monospace"}}>{h.pnlPct>=0?"+":""}{h.pnlPct.toFixed(2)}%</div><div style={{fontSize:9,color:h.pnl>=0?"#2A9D8F60":"#E6394660",fontFamily:"'JetBrains Mono',monospace"}}>{h.pnl>=0?"+":""}{fmtVal(h.pnl,cur)}</div></>}</div>
                <div style={{display:"flex",justifyContent:"flex-end"}}>{h.category!=="cash"&&<Spark up={h.pnlPct>=0} />}</div>
              </div>))}
            </div>);
          })}
        </>)}

        {/* ═══ TRANSACTIONS ═══ */}
        {page==="transactions"&&(<>
          <div style={{display:"flex",justifyContent:"space-between",alignItems:"center",marginBottom:12}}>
            <h2 style={{fontSize:16,fontWeight:700,color:"#EAEDF6"}}>Transaction History</h2>
            <button onClick={()=>setShowAddTx(true)} style={B("#2A9D8F","#0A0C12")}>+ Add Transaction</button>
          </div>
          <div style={{display:"flex",gap:10,alignItems:"center",marginBottom:10,padding:"8px 12px",borderRadius:10,background:"#0F1119",border:"1px solid #16182230"}}>
            <span style={{fontSize:10,color:"#4A4E60",fontWeight:600}}>Filter:</span>
            <div style={{display:"flex",alignItems:"center",gap:5}}><label style={{fontSize:10,color:"#6B7084"}}>From</label><input type="date" value={dateFrom} onChange={e=>{setDateFrom(e.target.value);setTxPage(1)}} style={{...I,width:130,padding:"4px 8px",fontSize:11}} /></div>
            <div style={{display:"flex",alignItems:"center",gap:5}}><label style={{fontSize:10,color:"#6B7084"}}>To</label><input type="date" value={dateTo} onChange={e=>{setDateTo(e.target.value);setTxPage(1)}} style={{...I,width:130,padding:"4px 8px",fontSize:11}} /></div>
            {(dateFrom||dateTo)&&<button onClick={()=>{setDateFrom("");setDateTo("");setTxPage(1)}} style={{...B("transparent","#E63946"),border:"1px solid #E6394630",fontSize:10,padding:"3px 8px"}}>Clear</button>}
            <div style={{marginLeft:"auto",display:"flex",alignItems:"center",gap:5}}><span style={{fontSize:10,color:"#4A4E60"}}>Per page:</span><select value={txPerPage} onChange={e=>{setTxPerPage(+e.target.value);setTxPage(1)}} style={{...I,width:55,padding:"3px 5px",fontSize:11}}>{[10,25,50,100].map(n=><option key={n} value={n}>{n}</option>)}</select></div>
          </div>
          <div style={{borderRadius:10,overflow:"hidden",background:"#0F1119",border:"1px solid #16182230"}}>
            <div style={{display:"grid",gridTemplateColumns:"82px 46px 74px 64px 70px 82px 82px 1fr",padding:"6px 10px",fontSize:10,color:"#3A3E50",textTransform:"uppercase",letterSpacing:"0.04em",fontWeight:600,borderBottom:"1px solid #14161F"}}><span>Date</span><span>Type</span><span>Category</span><span>Ticker</span><span style={{textAlign:"right"}}>Qty</span><span style={{textAlign:"right"}}>Price</span><span style={{textAlign:"right"}}>Total</span><span style={{paddingLeft:8}}>Notes</span></div>
            {txSlice.map((t,i)=>{const cat=CATEGORIES.find(c=>c.key===t.category);return(
              <div key={t.id} style={{display:"grid",gridTemplateColumns:"82px 46px 74px 64px 70px 82px 82px 1fr",padding:"7px 10px",alignItems:"center",borderBottom:i<txSlice.length-1?"1px solid #10121A":"none"}}>
                <span style={{fontSize:12,color:"#8B8FA3",fontFamily:"'JetBrains Mono',monospace"}}>{t.date}</span>
                <span style={{fontSize:10,fontWeight:700,padding:"2px 5px",borderRadius:4,width:"fit-content",background:t.type==="buy"?"#2A9D8F15":"#E6394615",color:t.type==="buy"?"#2A9D8F":"#E63946",textTransform:"uppercase"}}>{t.type}</span>
                <div style={{display:"flex",alignItems:"center",gap:4}}><div style={{width:6,height:6,borderRadius:2,background:cat?.color}} /><span style={{fontSize:11,color:"#6B7084"}}>{cat?.label}</span></div>
                <span style={{fontSize:12,fontWeight:600,color:"#EAEDF6"}}>{t.ticker}</span>
                <span style={{textAlign:"right",fontSize:12,color:"#C8CCD8",fontFamily:"'JetBrains Mono',monospace"}}>{t.qty.toLocaleString()}</span>
                <span style={{textAlign:"right",fontSize:11,color:"#8B8FA3",fontFamily:"'JetBrains Mono',monospace"}}>{fmtRp(t.price)}</span>
                <span style={{textAlign:"right",fontSize:11,color:"#EAEDF6",fontFamily:"'JetBrains Mono',monospace",fontWeight:600}}>{fmtRp(t.qty*t.price)}</span>
                <span style={{paddingLeft:8,fontSize:11,color:"#4A4E60",overflow:"hidden",textOverflow:"ellipsis",whiteSpace:"nowrap"}}>{t.notes||"—"}</span>
              </div>
            );})}
          </div>
          <div style={{display:"flex",justifyContent:"space-between",alignItems:"center",marginTop:8}}>
            <span style={{fontSize:10,color:"#3A3E50"}}>Showing {(txPage-1)*txPerPage+1}–{Math.min(txPage*txPerPage,filteredTx.length)} of {filteredTx.length}</span>
            <div style={{display:"flex",gap:4}}>
              <button onClick={()=>setTxPage(Math.max(1,txPage-1))} disabled={txPage===1} style={{...B("#16182280","#6B7084"),opacity:txPage===1?0.3:1,padding:"4px 10px"}}>← Prev</button>
              {Array.from({length:Math.min(txTotalPages,5)},(_,i)=>i+1).map(p=>(<button key={p} onClick={()=>setTxPage(p)} style={{...B(p===txPage?"#D4A84320":"#16182240",p===txPage?"#D4A843":"#4A4E60"),padding:"4px 10px",minWidth:30,fontSize:10,border:p===txPage?"1px solid #D4A84340":"1px solid transparent"}}>{p}</button>))}
              <button onClick={()=>setTxPage(Math.min(txTotalPages,txPage+1))} disabled={txPage===txTotalPages} style={{...B("#16182280","#6B7084"),opacity:txPage===txTotalPages?0.3:1,padding:"4px 10px"}}>Next →</button>
            </div>
          </div>
        </>)}
      </div>

      {/* ═══ ADD TRANSACTION MODAL ═══ */}
      {showAddTx&&(
        <div style={{position:"fixed",inset:0,background:"#00000080",backdropFilter:"blur(4px)",display:"flex",alignItems:"center",justifyContent:"center",zIndex:50}} onClick={()=>setShowAddTx(false)}>
          <div onClick={e=>e.stopPropagation()} style={{width:460,padding:24,borderRadius:14,background:"#12141D",border:"1px solid #1E2130",boxShadow:"0 20px 60px #00000060",maxHeight:"90vh",overflowY:"auto"}}>
            <div style={{fontSize:14,fontWeight:700,color:"#EAEDF6",marginBottom:16}}>Add Transaction</div>

            {/* Buy/Sell */}
            <div style={{marginBottom:14}}><label style={L}>Transaction Type</label>
              <div style={{display:"flex",gap:6}}>{["buy","sell"].map(t=>(<button key={t} onClick={()=>setTxForm({...txForm,type:t})} style={{flex:1,padding:"8px",borderRadius:7,border:`1.5px solid ${txForm.type===t?(t==="buy"?"#2A9D8F":"#E63946"):"#1E2130"}`,background:txForm.type===t?(t==="buy"?"#2A9D8F10":"#E6394610"):"transparent",color:txForm.type===t?(t==="buy"?"#2A9D8F":"#E63946"):"#4A4E60",fontSize:12,fontWeight:600,cursor:"pointer",fontFamily:"inherit",textTransform:"uppercase"}}>{t==="buy"?"🟢 Buy":"🔴 Sell"}</button>))}</div>
            </div>

            {/* Category */}
            <div style={{marginBottom:14}}><label style={L}>Asset Category</label>
              <div style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:5}}>{CATEGORIES.map(cat=>(<button key={cat.key} onClick={()=>setTxForm({type:txForm.type,category:cat.key,ticker:"",name:"",qty:"",price:"",notes:"",annualReturn:""})} style={{padding:"7px 6px",borderRadius:7,border:`1.5px solid ${txForm.category===cat.key?cat.color:"#1E2130"}`,background:txForm.category===cat.key?`${cat.color}10`:"transparent",color:txForm.category===cat.key?cat.color:"#4A4E60",fontSize:10,fontWeight:600,cursor:"pointer",fontFamily:"inherit",display:"flex",alignItems:"center",justifyContent:"center",gap:4}}><span>{cat.icon}</span> {cat.label}</button>))}</div>
            </div>

            {/* CRYPTO / IDX / US / GOLD — ticker search */}
            {["crypto","idx","us","gold"].includes(txForm.category)&&(
              <div style={{marginBottom:14}}>
                <label style={L}>{txForm.category==="gold"?"Gold Type":"Ticker / Symbol"} {["crypto","idx","us"].includes(txForm.category)&&"(type 3+ chars for suggestions)"}</label>
                <input type="text" placeholder={txForm.category==="crypto"?"e.g. BTC, ETH, BNB...":txForm.category==="idx"?"e.g. BBCA, TLKM, BBRI...":txForm.category==="us"?"e.g. AAPL, NVDA, VOO...":"e.g. Antam Gold..."} value={txForm.ticker} onChange={e=>setTxForm({...txForm,ticker:e.target.value})} style={I} />
                {txForm.ticker.length>=3&&["crypto","idx","us"].includes(txForm.category)&&(
                  <div style={{marginTop:4,borderRadius:7,background:"#181B27",border:"1px solid #1E2130",overflow:"hidden"}}>
                    <div style={{padding:"6px 12px",fontSize:10,color:"#4A4E60",borderBottom:"1px solid #14161F"}}><span style={{color:"#D4A843"}}>⟳</span> Searching "{txForm.ticker}"...</div>
                    <div style={{padding:"8px 12px",fontSize:12,color:"#EAEDF6",cursor:"pointer",display:"flex",justifyContent:"space-between",alignItems:"center"}} onClick={()=>setTxForm({...txForm,ticker:txForm.ticker.toUpperCase(),name:txForm.ticker.toUpperCase()})}>
                      <span><strong>{txForm.ticker.toUpperCase()}</strong> <span style={{color:"#4A4E60",fontSize:10}}>— Tap to select</span></span>
                      <span style={{fontSize:9,color:"#2A9D8F",fontWeight:600,padding:"2px 6px",background:"#2A9D8F15",borderRadius:4}}>SELECT</span>
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* CASH — IDR / USD */}
            {txForm.category==="cash"&&(
              <div style={{marginBottom:14}}><label style={L}>Currency</label>
                <div style={{display:"flex",gap:6}}>
                  {[{k:"IDR",l:"🇮🇩 IDR — Indonesian Rupiah",c:"#E63946"},{k:"USD",l:"🇺🇸 USD — US Dollar",c:"#457B9D"}].map(c=>(<button key={c.k} onClick={()=>setTxForm({...txForm,ticker:c.k,name:c.k==="IDR"?"Indonesian Rupiah":"US Dollar"})} style={{flex:1,padding:"10px 8px",borderRadius:7,border:`1.5px solid ${txForm.ticker===c.k?c.c:"#1E2130"}`,background:txForm.ticker===c.k?`${c.c}10`:"transparent",color:txForm.ticker===c.k?c.c:"#4A4E60",fontSize:11,fontWeight:600,cursor:"pointer",fontFamily:"inherit",textAlign:"center"}}>{c.l}</button>))}
                </div>
              </div>
            )}

            {/* BONDS — Name + Annual Return */}
            {txForm.category==="bonds"&&(<>
              <div style={{marginBottom:14}}><label style={L}>Bond Name / Series</label><input type="text" placeholder="e.g. ORI024, SR020, FR0098..." value={txForm.name} onChange={e=>setTxForm({...txForm,name:e.target.value,ticker:e.target.value})} style={I} /></div>
              <div style={{marginBottom:14}}><label style={L}>Annual Return (%)</label>
                <div style={{position:"relative"}}><input type="number" placeholder="e.g. 6.25" value={txForm.annualReturn} onChange={e=>setTxForm({...txForm,annualReturn:e.target.value})} style={{...I,paddingRight:30}} /><span style={{position:"absolute",right:10,top:"50%",transform:"translateY(-50%)",fontSize:12,color:"#4A4E60",fontWeight:600}}>%</span></div>
                <div style={{fontSize:9,color:"#3A3E50",marginTop:4}}>The coupon/annual interest rate of this bond</div>
              </div>
            </>)}

            {/* Qty + Price */}
            <div style={{display:"grid",gridTemplateColumns:txForm.category==="cash"?"1fr":"1fr 1fr",gap:10,marginBottom:14}}>
              <div><label style={L}>{txForm.category==="cash"?"Amount":txForm.category==="gold"?"Quantity (gram)":txForm.category==="bonds"?"Face Value (Rp)":"Quantity"}</label><input type="number" placeholder={txForm.category==="cash"?"e.g. 50000000":txForm.category==="bonds"?"e.g. 150000000":"0"} value={txForm.qty} onChange={e=>setTxForm({...txForm,qty:e.target.value})} style={I} /></div>
              {txForm.category!=="cash"&&(<div><label style={L}>{txForm.category==="bonds"?"Purchase Price (Rp)":txForm.category==="gold"?"Price per gram (Rp)":txForm.category==="us"?"Price per unit (USD)":"Price per unit"}</label><input type="number" placeholder="0" value={txForm.price} onChange={e=>setTxForm({...txForm,price:e.target.value})} style={I} /></div>)}
            </div>

            {/* Notes (not for bonds) */}
            {txForm.category!=="bonds"&&(<div style={{marginBottom:18}}><label style={L}>Notes (optional)</label><input type="text" placeholder="Any notes..." value={txForm.notes} onChange={e=>setTxForm({...txForm,notes:e.target.value})} style={I} /></div>)}

            <div style={{display:"flex",gap:8,justifyContent:"flex-end"}}>
              <button onClick={()=>{setShowAddTx(false);resetForm()}} style={B("transparent","#6B7084")}>Cancel</button>
              <button onClick={handleAddTx} style={{...B(txForm.type==="buy"?"#2A9D8F":"#E63946","#fff"),opacity:canSubmit?1:0.4}}>{txForm.type==="buy"?"✓ Record Buy":"✓ Record Sell"}</button>
            </div>
          </div>
        </div>
      )}
      <div style={{textAlign:"center",padding:"20px 0 10px",fontSize:9,color:"#1A1D2A"}}>PortfolioIQ · Serverless · Telegram Alerts · $0/mo</div>
    </div>
  );
}
