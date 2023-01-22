package main

import (
    "encoding/binary"
    "fmt"
    "github.io/vsverchkov/d-cache/cache"
    "github.io/vsverchkov/d-cache/client"
    "github.io/vsverchkov/d-cache/proto"
    "io"
    "log"
    "net"
    "time"
)

type ServerOpts struct {
    ListenAddr string
    IsLeader bool
    LeaderAddr string
}

type Server struct {
    ServerOpts
    nodes map[*client.Client]struct{}
    cache cache.Cacher
}

func NewServer(opts ServerOpts, c cache.Cacher) *Server {
    return &Server{
        ServerOpts: opts,
        nodes: make(map[*client.Client]struct{}),
        cache: c,
    }
}

func (s *Server) Start() error {
    ln, err := net.Listen("tcp", s.ListenAddr)
    if err != nil {
        return fmt.Errorf("listen error: %s", err)
    }

    if !s.IsLeader && len(s.ListenAddr) != 0 {
        go func() {
            if err := s.deadLeader(); err != nil {
                log.Println(err)
            }
        }()
    }

    log.Printf("server starting on port [%s]\n", s.ListenAddr)

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Printf("accept error: %s", err)
            continue
        }
        go s.handleConn(conn)
    }
}

func (s *Server) deadLeader() error {
    conn, err := net.Dial("tcp", s.LeaderAddr)
    if err != nil {
        return fmt.Errorf("failed to deal leader [%s]", s.LeaderAddr)
    }

    log.Println("connected to leader:", s.LeaderAddr)

    binary.Write(conn, binary.LittleEndian, proto.CmdJoin)

    s.handleConn(conn)
    return nil
}

func (s *Server) handleConn(conn net.Conn) {
    defer conn.Close()

    for {
        cmd, err := proto.ParseCommand(conn)
        if err != nil {
            if err == io.EOF {
                break
            }
            log.Printf("parse command error: %s\n", err)
            break
        }

        go s.handleCommand(conn, cmd)
    }
}

func (s *Server) handleCommand(conn net.Conn, cmd any) {
    switch t := cmd.(type) {
    case *proto.CommandSet:
        s.handleSetCommand(conn, t)
    case *proto.CommandGet:
        s.handleGetCommand(conn, t)
    case *proto.CommandJoin:
        s.handleJoinCommand(conn, t)
    }
}

func (s *Server) handleSetCommand(conn net.Conn, cmd *proto.CommandSet) error {
    go func() {
        for node := range s.nodes {
            err := node.Set(cmd.Key, cmd.Value, cmd.TTL)
            if err != nil {
                log.Println("forward to node error:", err)
            }
        }
    }()

    resp := proto.ResponseSet{}
    if err := s.cache.Set(cmd.Key, cmd.Value, time.Duration(cmd.TTL)); err != nil {
        resp.Status = proto.StatusError
        conn.Write(resp.Bytes())
        return err
    }

    resp.Status = proto.StatusOK
    _, err := conn.Write(resp.Bytes())
    return err
}

func (s *Server) handleGetCommand(conn net.Conn, cmd *proto.CommandGet) error {
    resp := proto.ResponseGet{}

    value, err := s.cache.Get(cmd.Key)
    if err != nil {
        resp.Status = proto.StatusError
        conn.Write(resp.Bytes())
        return err
    }

    if len(value) == 0 {
        resp.Status = proto.StatusKeyNotFound
        conn.Write(resp.Bytes())
        return nil
    }

    resp.Status = proto.StatusOK
    resp.Value = value
    _, err = conn.Write(resp.Bytes())

    return err
}

func (s *Server) handleJoinCommand(conn net.Conn, cmd *proto.CommandJoin) error {
    fmt.Println("node just joined the cluster:", conn.RemoteAddr())

    s.nodes[client.NewFromConn(conn)] = struct{}{}

    return nil
}