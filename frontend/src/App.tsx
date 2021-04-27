import React, { useState, MouseEvent, useCallback, useRef, FC } from "react";
import PackagePicker from "./PackagePicker";
import ReactDiffViewer from 'react-diff-viewer';
import { DndProvider, DragPreviewImage, DropTargetMonitor, useDrop, XYCoord } from 'react-dnd'
import { HTML5Backend } from 'react-dnd-html5-backend'
import { DraggableList } from "./DraggableList"
import update from 'immutability-helper'

declare var Prism: any;

interface FileInfo {
  [key: string]: string
}

export const ItemTypes = {
  TEST: 'test'
}

export default function App() {
  const [tests, setTests] = useState([]);
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [landingPage, setLandingPage] = useState(true);
  const [activeTest, setActiveTest] = useState(-1);
  const [activePkg, setActivePkg] = useState("");

  const fetchTestNames = async (pkg: string) => {
    return fetch('http://localhost:3000/listTests?pkg=' + pkg)
      .then(r => r.json())
  }

  const fetchFiles = async (pkg: string, testNames: string[]) => {
    return fetch('http://localhost:3000/listFiles', {
      method: 'POST',
      body: JSON.stringify({
        pkg,
        tests: testNames,
      })
    })
      .then(r => r.json())
  }

  const highlightSyntax = (str: string) => (
    <pre
      style={{ display: 'inline' }}
      dangerouslySetInnerHTML={{
        __html: Prism.highlight(str || "", Prism.languages.clike, 'javascript'),
      }}
    />
  );

  const filesView = () => {
    if (activeTest == -1) return 'Select a test to begin';
    const previousContents = activeTest === 0 ? {} : files[activeTest - 1];
    const contents = files[activeTest];

    const out = [];
    for (const [name, content] of Object.entries(contents)) {
      const previousContent = previousContents[name] || '';
      console.log("nme: , name: ", name)
      out.push(
        <div key={name} className="File">
          <ReactDiffViewer
            oldValue={atob(previousContent)}
            newValue={atob(content)}
            splitView={false}
            leftTitle={name}
            showDiffOnly={Object.keys(contents).length > 1 ? true : false}
            renderContent={highlightSyntax}
          />
        </div>
      )
    }

    return out
  }


  async function handleSubmit(pkg: string) {
    setActivePkg(pkg)
    const testNames = await fetchTestNames(pkg);
    setTests(testNames);
  }

  async function handleGenerateLogs() {
    fetchFiles(activePkg, tests).then(setFiles)
  }

  // Landing Page
  if (tests.length === 0 && files.length === 0) {
    return (
      <div className="LandingPage">
        <PackagePicker onSubmit={handleSubmit} />
      </div>
    )
  }

  // Pkg selected, tests not ordered
  if (files.length === 0) {
    return (
      <div className="TestOrdering">
        <PackagePicker onSubmit={handleSubmit} />
        <div>
          Order your tests man
        </div>
        <DndProvider backend={HTML5Backend}>
          <DraggableList setItems={setTests} items={tests} />
        </DndProvider>
        <button onClick={handleGenerateLogs}>Generate Log</button>
      </div>
    )
  }

  return (
    <div>
      <PackagePicker onSubmit={handleSubmit} />
      <br /><br />

      <div className="Page">
        <div className="Tests">
          <div>
            Tests
            ---------------
          </div>
          {tests.map((t, i) => <button key={i} name={t} className={i === activeTest ? 'is-active' : ''} onClick={() => setActiveTest(i)}>{t}</button>)}
          <button className={tests.length == activeTest ? 'is-active' : ''} onClick={() => setActiveTest(tests.length)}>Final</button>
        </div>
        <div className="Files">
          {filesView()}
        </div>
      </div>
    </div>
  )
}
