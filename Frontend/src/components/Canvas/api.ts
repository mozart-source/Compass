import { gql } from '@apollo/client'

export const GET_CANVASES = gql`
  query GetCanvases {
    listCanvases {
      success
      data {
        id
        title
        description
        tags
        nodes {
          id
          type
          data
          position {
            x
            y
          }
          style
          label
        }
        edges {
          id
          source
          target
          type
          data
          style
          label
        }
        updatedAt
        createdAt
      }
    }
  }
`

export const GET_CANVAS = gql`
  query GetCanvas($id: ID!) {
    getCanvas(id: $id) {
      success
      data {
        id
        title
        description
        tags
        nodes {
          id
          type
          data
          position {
            x
            y
          }
          style
          label
        }
        edges {
          id
          source
          target
          type
          data
          style
          label
        }
        updatedAt
        createdAt
      }
    }
  }
`

export const CREATE_CANVAS = gql`
  mutation CreateCanvas($input: CanvasInput!) {
    createCanvas(input: $input) {
      success
      data {
        id
        title
        description
        tags
        nodes {
          id
        }
        edges {
          id
        }
        updatedAt
        createdAt
      }
    }
  }
`

export const UPDATE_CANVAS = gql`
  mutation UpdateCanvas($id: ID!, $input: CanvasInput!) {
    updateCanvas(id: $id, input: $input) {
      success
      data {
        id
        title
        description
        tags
        nodes {
          id
        }
        edges {
          id
        }
        updatedAt
      }
    }
  }
`

export const DELETE_CANVAS = gql`
  mutation DeleteCanvas($id: ID!) {
    deleteCanvas(id: $id) {
      success
      data {
        id
      }
    }
  }
`

export const CREATE_NODE = gql`
  mutation CreateNode($input: CanvasNodeInput!) {
    createCanvasNode(input: $input) {
      success
      data {
        id
        type
        data
        position {
          x
          y
        }
        style
        label
      }
    }
  }
`

export const UPDATE_NODE = gql`
  mutation UpdateNode($id: ID!, $input: CanvasNodeInput!) {
    updateCanvasNode(id: $id, input: $input) {
      success
      data {
        id
        type
        data
        position {
          x
          y
        }
        style
        label
      }
    }
  }
`

export const DELETE_NODE = gql`
  mutation DeleteNode($id: ID!) {
    deleteCanvasNode(id: $id) {
      success
      data {
        id
      }
    }
  }
`

export const CREATE_EDGE = gql`
  mutation CreateEdge($input: CanvasEdgeInput!) {
    createCanvasEdge(input: $input) {
      success
      data {
        id
        source
        target
        type
        data
        style
        label
      }
    }
  }
`

export const UPDATE_EDGE = gql`
  mutation UpdateEdge($id: ID!, $input: CanvasEdgeInput!) {
    updateCanvasEdge(id: $id, input: $input) {
      success
      data {
        id
        source
        target
        type
        data
        style
        label
      }
    }
  }
`

export const DELETE_EDGE = gql`
  mutation DeleteEdge($id: ID!) {
    deleteCanvasEdge(id: $id) {
      success
      data {
        id
      }
    }
  }
`
