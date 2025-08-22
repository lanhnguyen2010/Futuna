import React, { useEffect, useState } from 'react';
import dynamic from 'next/dynamic';

const ReactTabulator = dynamic(() => import('react-tabulator'), { ssr: false });

export default function AnalysisTable() {
  const [rows, setRows] = useState([]);
  const [strategies, setStrategies] = useState([]);
  const [active, setActive] = useState('All');
  const api = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  useEffect(() => {
    fetch(`${api}/api/analysis`)
      .then((res) => res.json())
      .then((data) => {
        setRows(data);
        if (data.length > 0) {
          let s = data[0].strategies;
          if (typeof s === 'string') {
            try {
              s = JSON.parse(s);
            } catch (e) {
              console.error(e);
              s = [];
            }
          }
          setStrategies(Array.isArray(s) ? s : []);
        }
      })
      .catch(console.error);
  }, [api]);

  const colorFormatter = (cell) => {
    const value = cell.getValue();
    const field = cell.getField();
    if (value.startsWith('ACCUMULATE')) {
      cell.getElement().style.color = 'green';
    } else if (value.startsWith('AVOID')) {
      cell.getElement().style.color = 'red';
    } else {
      cell.getElement().style.color = 'orange';
    }
    let confField = null;
    if (field === 'short_term') confField = 'short_confidence';
    if (field === 'long_term') confField = 'long_confidence';
    if (confField) {
      const conf = cell.getRow().getData()[confField];
      if (conf !== undefined) {
        cell.getElement().setAttribute('title', conf + '%');
      }
    }
    return value;
  };

  const columns = [
    { title: 'Ticker', field: 'ticker', hozAlign: 'left' },
    { title: 'Short', field: 'short_term', formatter: colorFormatter },
    { title: 'Long', field: 'long_term', formatter: colorFormatter },
    { title: 'Overall', field: 'overall', formatter: colorFormatter }
  ];

  const extractTickers = (note) => {
    const tickerSet = new Set();
    const matches = note.match(/\b[A-Z]{3,4}\b/g) || [];
    matches.forEach((t) => {
      if (rows.some((r) => r.ticker === t)) {
        tickerSet.add(t);
      }
    });
    return Array.from(tickerSet);
  };

  const dataForTab = () => {
    if (active === 'All') return rows;
    const strat = strategies.find((s) => s.name === active);
    if (!strat) return [];
    const tickers = extractTickers(strat.note);
    return rows.filter((r) => tickers.includes(r.ticker));
  };

  return (
    <div>
      <ul className="tabs">
        <li className={active === 'All' ? 'active' : ''} onClick={() => setActive('All')}>All</li>
        {strategies.map((s) => (
          <li key={s.name} className={active === s.name ? 'active' : ''} onClick={() => setActive(s.name)}>
            {s.name}
          </li>
        ))}
      </ul>
      {active !== 'All' && (
        <div className="strategy-info">
          <p><strong>Stance:</strong> {strategies.find((s) => s.name === active)?.stance}</p>
          <p><strong>Note:</strong> {strategies.find((s) => s.name === active)?.note}</p>
        </div>
      )}
      <ReactTabulator data={dataForTab()} columns={columns} layout="fitData" />
      <style jsx>{`
        .tabs {
          list-style: none;
          padding: 0;
          display: flex;
          gap: 1rem;
          cursor: pointer;
        }
        .tabs li {
          padding: 0.5rem 1rem;
          border-bottom: 2px solid transparent;
        }
        .tabs li.active {
          border-bottom-color: #000;
        }
        .strategy-info {
          margin: 1rem 0;
        }
      `}</style>
    </div>
  );
}
