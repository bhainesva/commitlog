import React, { useState, MouseEvent, useRef, FC } from "react";
import classNames from 'classnames';
import PackagePicker from "./PackagePicker";
import ReactDiffViewer from 'react-diff-viewer';
import { DndProvider, DropTargetMonitor, useDrop, XYCoord } from 'react-dnd'
import { HTML5Backend } from 'react-dnd-html5-backend'

import { useDrag } from 'react-dnd'

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
    const testNames = await fetchTestNames(pkg);
    setTests(testNames);
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
          {tests.map((t, i) => <Test index={i} moveCard={i => console.log(i)} id={i} key={i} name={t} isActive={i === activeTest} onClick={() => setActiveTest(i)} />)}
        </DndProvider>
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
          {/* {tests.map((t, i) => <Test key={i} name={t} isActive={i === activeTest} onClick={() => setActiveTest(i)} />)} */}
          <button className={tests.length == activeTest ? 'is-active' : ''} onClick={() => setActiveTest(tests.length)}>Final</button>
        </div>
        <div className="Files">
          {filesView()}
        </div>
      </div>
    </div>
  )
}

interface DragItem {
  index: number
  id: string
  type: string
}

export interface TestProps {
  id: any
  index: number
  name: string,
  isActive: boolean,
  onClick: (e: MouseEvent<HTMLButtonElement>) => void,
  moveCard: (dragIndex: number, hoverIndex: number) => void
}

const Test: FC<TestProps> = ({ id, index, name, moveCard, isActive }) => {
  const ref = useRef<HTMLDivElement>(null);
  const [{ handlerId }, drop] = useDrop(() => ({
    accept: ItemTypes.TEST,
    collect: (monitor) => ({
      handlerId: monitor.getHandlerId(),
    }),
    hover(item: DragItem, monitor: DropTargetMonitor) {
      if (!ref.current) {
        return
      }

      const dragIndex = item.index;
      const hoverIndex = index;

      if (dragIndex === hoverIndex) {
        return
      }

      const hoverBoundingRect = ref.current?.getBoundingClientRect()

      // Get vertical middle
      const hoverMiddleY =
        (hoverBoundingRect.bottom - hoverBoundingRect.top) / 2

      // Determine mouse position
      const clientOffset = monitor.getClientOffset()

      // Get pixels to the top
      const hoverClientY = (clientOffset as XYCoord).y - hoverBoundingRect.top

      // Only perform the move when the mouse has crossed half of the items height
      // When dragging downwards, only move when the cursor is below 50%
      // When dragging upwards, only move when the cursor is above 50%

      // Dragging downwards
      if (dragIndex < hoverIndex && hoverClientY < hoverMiddleY) {
        return
      }

      // Dragging upwards
      if (dragIndex > hoverIndex && hoverClientY > hoverMiddleY) {
        return
      }

      // Time to actually perform the action
      moveCard(dragIndex, hoverIndex)

      // Note: we're mutating the monitor item here!
      // Generally it's better to avoid mutations,
      // but it's good here for the sake of performance
      // to avoid expensive index searches.
      item.index = hoverIndex
    }
  }))

  const [{ isDragging }, drag] = useDrag({
    type: ItemTypes.TEST,
    item: () => {
      return { id, index }
    },
    collect: (monitor: any) => ({
      isDragging: monitor.isDragging(),
    }),
  })

  const opacity = isDragging ? 0 : 1;
  drag(drop(ref));

  return (
    <div ref={drag}
      style={{
        opacity,
        fontSize: 25,
        fontWeight: 'bold',
        cursor: 'move',
      }}>
      TEST WRAPPER
      <br />
      <button className={classNames({ 'Test': true, 'is-active': isActive })}>
        {name}
      </button>
    </div>
  )
}