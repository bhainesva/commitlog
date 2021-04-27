import React from "react"
import { FC, useState, useCallback } from 'react'
import { Card } from './Card'
import update from 'immutability-helper'

const style = {
  width: 400,
}

export interface Item {
  id: number
  text: string
}

export interface ContainerState {
  cards: Item[]
}

export interface DraggableListProps {
    items: string[]
    setItems: ([]) => void
}

export const DraggableList: FC<DraggableListProps> = ({items, setItems}) => {
  {
    const cards = [...items];

    const moveItem = (dragIndex: number, hoverIndex: number) => {
       const dragCard = cards[dragIndex]
       setItems(
          update(cards, {
            $splice: [
              [dragIndex, 1],
              [hoverIndex, 0, dragCard],
            ],
          }),
        )
      }


    const renderCard = (card: string, index: number) => {
      return (
        <Card
          key={card}
          index={index}
          id={card}
          text={card}
          moveCard={moveItem}
        />
      )
    }

    return (
      <>
        <div style={style}>{cards.map((card, i) => renderCard(card, i))}</div>
      </>
    )
  }
}
