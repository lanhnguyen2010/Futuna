import React, { useEffect, useState } from "react";
import dynamic from "next/dynamic";

const ReactTabulator = dynamic(
  () => import("react-tabulator").then((mod) => mod.ReactTabulator),
  { ssr: false },
);

export default function AnalysisTable() {
  const [rows, setRows] = useState([]);
  const [dates, setDates] = useState([]);
  const [strategies, setStrategies] = useState([]);
  const [active, setActive] = useState("All");
  const [date, setDate] = useState("");
  const [prevDate, setPrevDate] = useState("");
  const [dateError, setDateError] = useState("");
  const [search, setSearch] = useState("");
  const [sources, setSources] = useState([]);
  const api = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const splitField = (str) => {
    if (!str) return ["", ""];
    const idx = str.indexOf(" - ");
    if (idx === -1) return [str, ""];
    return [str.slice(0, idx), str.slice(idx + 3)];
  };

  useEffect(() => {
    const url = date ? `${api}/api/analysis?date=${date}` : `${api}/api/analysis`;
    fetch(url)
      .then((res) => res.json())
      .then((data) => {
        const stratMap = new Map();
        const parsed = data.map((r) => {
          let s = r.strategies;
          if (typeof s === "string") {
            try {
              s = JSON.parse(s);
            } catch (e) {
              console.error(e);
              s = [];
            }
          }
          (s || []).forEach((st) => {
            if (!stratMap.has(st.name)) stratMap.set(st.name, st.name);
          });
          r.strategies = Array.isArray(s) ? s : [];
          if (typeof r.sources === "string") {
            try {
              r.sources = JSON.parse(r.sources);
            } catch {
              r.sources = [];
            }
          }
          const [shortRec, shortReason] = splitField(r.short_term);
          const [longRec, longReason] = splitField(r.long_term);
          const [overallRec, overallReason] = splitField(r.overall);
          r.short_term = shortRec;
          r.short_reason = shortReason;
          r.long_term = longRec;
          r.long_reason = longReason;
          r.overall = overallRec;
          r.overall_reason = overallReason;
          return r;
        });
        setRows(parsed);
        setStrategies(Array.from(stratMap.keys()));
        if (parsed.length > 0) {
          setSources(Array.isArray(parsed[0].sources) ? parsed[0].sources : []);
        }
      })
      .catch(console.error);
  }, [api, date]);

  // fetch available dates for analyses
  useEffect(() => {
    fetch(`${api}/api/dates`)
      .then((res) => res.json())
      .then((d) => {
        const list = d || [];
        setDates(list);
        if (list && list.length > 0) {
          // default to the first (most recent) date
          setDate(list[0]);
          setPrevDate(list[0]);
        } else {
          const today = new Date().toISOString().slice(0, 10);
          setDate(today);
          setPrevDate(today);
        }
      })
      .catch(() => {
        const today = new Date().toISOString().slice(0, 10);
        setDate(today);
        setPrevDate(today);
      });
  }, [api]);

  const colorFormatter = (cell) => {
    const raw = cell.getValue();
    const value =
      typeof raw === "string" ? raw : raw == null ? "" : String(raw);
    const field = cell.getField();
    if (value.startsWith("ACCUMULATE")) {
      cell.getElement().style.color = "green";
    } else if (value.startsWith("AVOID")) {
      cell.getElement().style.color = "red";
    } else {
      cell.getElement().style.color = "orange";
    }
    let confField = null;
    if (field === "short_term") confField = "short_confidence";
    if (field === "long_term") confField = "long_confidence";
    if (field === "overall") confField = "overall_confidence";
    if (confField) {
      const conf = cell.getRow().getData()[confField];
      if (conf !== undefined) {
        cell.getElement().setAttribute("title", conf + "%");
      }
    }
    return value;
  };

  const baseColumns = [
    { title: "Ticker", field: "ticker", hozAlign: "left" },
    { title: "Short Term", field: "short_term", formatter: colorFormatter },
    { title: "Short Details", field: "short_reason" },
    { title: "Long Term", field: "long_term", formatter: colorFormatter },
    { title: "Long Term Details", field: "long_reason" },
    { title: "Overall", field: "overall", formatter: colorFormatter },
  ];

  const columns =
    active === "All"
      ? baseColumns
      : [
          ...baseColumns,
          { title: "Stance", field: "strategy_stance" },
          { title: "Note", field: "strategy_note" },
        ];

  const dataForTab = () => {
    const filtered = rows.filter((r) => {
      if (!search) return true;
      return r.ticker.toLowerCase().includes(search.toLowerCase());
    });
    if (active === "All") return filtered;
    return filtered.reduce((acc, r) => {
      const st = r.strategies.find((s) => s.name === active);
      if (st)
        acc.push({ ...r, strategy_stance: st.stance, strategy_note: st.note });
      return acc;
    }, []);
  };

  return (
    <div>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: "1rem",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "0.5rem", flexDirection: "column" }}>
          <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
            <label htmlFor="analysis-date">Date:</label>
            <input
              id="analysis-date"
              type="date"
              value={date}
              onChange={(e) => {
                const v = e.target.value;
                // if we have a list of available dates, only allow selecting from that list
                if (dates.length > 0 && !dates.includes(v)) {
                  // revert to previous valid date and show a short error
                  setDate(prevDate);
                  setDateError("No data for selected date");
                  setTimeout(() => setDateError(""), 3000);
                  return;
                }
                setPrevDate(date);
                setDate(v);
              }}
            />
          </div>
          <div style={{ fontSize: 12, color: "#666" }}>
            Available dates: {dates.length > 0 ? dates.join(", ") : "none"}
          </div>
          {dateError && (
            <div style={{ color: "red", fontSize: 12 }}>{dateError}</div>
          )}
        </div>
        <ul className="tabs">
          <li
            className={active === "All" ? "active" : ""}
            onClick={() => setActive("All")}
          >
            All
          </li>
          {strategies.map((name) => (
            <li
              key={name}
              className={active === name ? "active" : ""}
              onClick={() => setActive(name)}
            >
              {name}
            </li>
          ))}
        </ul>
        <div className="search">
          <label htmlFor="ticker-search" style={{ marginRight: 8 }}>
            Search Ticker:
          </label>
          <input
            id="ticker-search"
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="e.g. VCI"
          />
        </div>
      </div>
      <ReactTabulator data={dataForTab()} columns={columns} layout="fitData" />
      {sources.length > 0 && (
        <div className="sources">
          <h3>Nguồn</h3>
          <ul>
            {sources.map((s) => (
              <li key={s}>
                <a href={s} target="_blank" rel="noreferrer">
                  {s}
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}
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
        .sources {
          margin-top: 1rem;
        }
      `}</style>
    </div>
  );
}
