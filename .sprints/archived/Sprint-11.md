# Sprint 11: Monitoring & Analytics Platform

**Duration:** 2 weeks  
**Theme:** Operational Intelligence & Performance Monitoring  
**Sprint Goals:** Build comprehensive monitoring and analytics platform for operational insights

## Objectives

1. **Real-Time Metrics**: System performance and business metrics collection
2. **Analytics Dashboard**: Interactive dashboards for tactical and operational insights
3. **Alerting System**: Proactive monitoring with intelligent alerting
4. **Data Pipeline**: Stream processing for real-time analytics
5. **Reporting Engine**: Automated report generation for stakeholders

## User Stories

### Epic: Operational Intelligence Platform

**US-11.1: Real-Time System Monitoring**
```
As a system administrator
I want comprehensive real-time monitoring of system performance
So that I can ensure optimal system operation and quickly identify issues
```

**Acceptance Criteria:**
- CPU, memory, disk, and network monitoring
- Application performance metrics (response time, throughput)
- Database performance monitoring
- Connection pool and resource utilization tracking
- Custom metric collection and visualization

**US-11.2: Tactical Operations Dashboard**
```
As an operations commander
I want real-time tactical dashboards showing force disposition
So that I can make informed decisions based on current situational awareness
```

**Acceptance Criteria:**
- Real-time unit positions and status on tactical map
- Communication flow analysis and network health
- Mission progress tracking and milestone visualization
- Resource allocation and utilization metrics
- Threat detection and alert correlation

**US-11.3: Business Intelligence Analytics**
```
As a senior leader
I want analytical reports on system usage and operational effectiveness
So that I can make strategic decisions about resource allocation and training
```

**Acceptance Criteria:**
- User activity and engagement analytics
- System utilization trends and forecasting
- Mission effectiveness metrics
- Training and exercise analytics
- Cost and resource optimization recommendations

**US-11.4: Proactive Alerting System**
```
As an operations center analyst
I want intelligent alerting for system issues and tactical events
So that I can respond quickly to critical situations
```

**Acceptance Criteria:**
- Multi-level alerting with escalation policies
- Anomaly detection for unusual patterns
- Correlation engine for related events
- Integration with external notification systems
- Alert fatigue reduction through intelligent filtering

**US-11.5: Automated Reporting**
```
As a program manager
I want automated generation of operational and compliance reports
So that I can meet reporting requirements without manual effort
```

**Acceptance Criteria:**
- Scheduled report generation and distribution
- Customizable report templates and formats
- Compliance reporting for security and audit requirements
- Executive dashboard summaries
- Data export capabilities for external analysis

## Technical Implementation

### Metrics Collection System

**Metrics Manager**
```go
// pkg/metrics/manager.go
package metrics

import (
    "context"
    "sync"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsManager struct {
    registry    prometheus.Registerer
    collectors  map[string]Collector
    config      *Config
    logger      Logger
    mu          sync.RWMutex
}

type Config struct {
    Enabled             bool          `yaml:"enabled"`
    CollectionInterval  time.Duration `yaml:"collection_interval"`
    RetentionPeriod     time.Duration `yaml:"retention_period"`
    
    // Prometheus configuration
    PrometheusEnabled   bool          `yaml:"prometheus_enabled"`
    PrometheusPort      int           `yaml:"prometheus_port"`
    
    // InfluxDB configuration  
    InfluxDBEnabled     bool          `yaml:"influxdb_enabled"`
    InfluxDBURL         string        `yaml:"influxdb_url"`
    InfluxDBDatabase    string        `yaml:"influxdb_database"`
    InfluxDBUsername    string        `yaml:"influxdb_username"`
    InfluxDBPassword    string        `yaml:"influxdb_password"`
    
    // Custom metrics
    CustomMetrics       []CustomMetric `yaml:"custom_metrics"`
}

type CustomMetric struct {
    Name        string            `yaml:"name"`
    Type        string            `yaml:"type"` // counter, gauge, histogram
    Description string            `yaml:"description"`
    Labels      []string          `yaml:"labels"`
    Help        string            `yaml:"help"`
}

// Core system metrics
var (
    // Connection metrics
    ActiveConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "gotak_active_connections",
            Help: "Number of active client connections",
        },
        []string{"protocol", "type"},
    )
    
    // Message metrics
    MessagesProcessed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "gotak_messages_processed_total",
            Help: "Total number of messages processed",
        },
        []string{"type", "status"},
    )
    
    MessageProcessingDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "gotak_message_processing_duration_seconds",
            Help: "Time spent processing messages",
            Buckets: prometheus.DefBuckets,
        },
        []string{"type"},
    )
    
    // User metrics
    ActiveUsers = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "gotak_active_users",
            Help: "Number of active users",
        },
        []string{"role", "group"},
    )
    
    UserActions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "gotak_user_actions_total",
            Help: "Total number of user actions",
        },
        []string{"action", "user_role"},
    )
    
    // Mission metrics
    ActiveMissions = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "gotak_active_missions",
            Help: "Number of active missions",
        },
        []string{"classification", "priority"},
    )
    
    MissionTasks = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "gotak_mission_tasks",
            Help: "Number of mission tasks by status",
        },
        []string{"status", "priority"},
    )
    
    // System performance
    SystemCPUUsage = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "gotak_system_cpu_usage_percent",
            Help: "System CPU usage percentage",
        },
    )
    
    SystemMemoryUsage = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "gotak_system_memory_usage_bytes",
            Help: "System memory usage in bytes",
        },
    )
    
    DatabaseConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "gotak_database_connections",
            Help: "Number of database connections",
        },
        []string{"state"},
    )
    
    // Federation metrics
    FederationConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "gotak_federation_connections",
            Help: "Number of federation connections",
        },
        []string{"server", "status"},
    )
    
    FederationMessages = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "gotak_federation_messages_total",
            Help: "Total federation messages sent/received",
        },
        []string{"direction", "type", "server"},
    )
)

func NewMetricsManager(config *Config, logger Logger) *MetricsManager {
    return &MetricsManager{
        registry:   prometheus.DefaultRegisterer,
        collectors: make(map[string]Collector),
        config:     config,
        logger:     logger,
    }
}

func (mm *MetricsManager) Start(ctx context.Context) error {
    if !mm.config.Enabled {
        mm.logger.Info("Metrics collection disabled")
        return nil
    }
    
    // Register custom metrics
    for _, metric := range mm.config.CustomMetrics {
        if err := mm.registerCustomMetric(metric); err != nil {
            mm.logger.Error("Failed to register custom metric", "metric", metric.Name, "error", err)
        }
    }
    
    // Start system metrics collector
    systemCollector := NewSystemMetricsCollector(mm.logger)
    mm.registerCollector("system", systemCollector)
    
    // Start database metrics collector
    dbCollector := NewDatabaseMetricsCollector(mm.logger)
    mm.registerCollector("database", dbCollector)
    
    // Start application metrics collector
    appCollector := NewApplicationMetricsCollector(mm.logger)
    mm.registerCollector("application", appCollector)
    
    // Start collection routine
    go mm.collectMetrics(ctx)
    
    mm.logger.Info("Metrics manager started")
    return nil
}

func (mm *MetricsManager) collectMetrics(ctx context.Context) {
    ticker := time.NewTicker(mm.config.CollectionInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            mm.runCollection()
        }
    }
}

func (mm *MetricsManager) runCollection() {
    mm.mu.RLock()
    collectors := make([]Collector, 0, len(mm.collectors))
    for _, collector := range mm.collectors {
        collectors = append(collectors, collector)
    }
    mm.mu.RUnlock()
    
    for _, collector := range collectors {
        go func(c Collector) {
            if err := c.Collect(); err != nil {
                mm.logger.Error("Failed to collect metrics", "collector", c.Name(), "error", err)
            }
        }(collector)
    }
}

func (mm *MetricsManager) RecordUserAction(action, userRole string) {
    UserActions.WithLabelValues(action, userRole).Inc()
}

func (mm *MetricsManager) RecordMessageProcessed(messageType, status string, duration time.Duration) {
    MessagesProcessed.WithLabelValues(messageType, status).Inc()
    MessageProcessingDuration.WithLabelValues(messageType).Observe(duration.Seconds())
}

func (mm *MetricsManager) UpdateActiveUsers(role, group string, count int) {
    ActiveUsers.WithLabelValues(role, group).Set(float64(count))
}

type Collector interface {
    Name() string
    Collect() error
}

type SystemMetricsCollector struct {
    logger Logger
}

func NewSystemMetricsCollector(logger Logger) *SystemMetricsCollector {
    return &SystemMetricsCollector{logger: logger}
}

func (smc *SystemMetricsCollector) Name() string {
    return "system"
}

func (smc *SystemMetricsCollector) Collect() error {
    // Collect CPU usage
    cpuUsage, err := getCPUUsage()
    if err != nil {
        return fmt.Errorf("failed to get CPU usage: %w", err)
    }
    SystemCPUUsage.Set(cpuUsage)
    
    // Collect memory usage
    memUsage, err := getMemoryUsage()
    if err != nil {
        return fmt.Errorf("failed to get memory usage: %w", err)
    }
    SystemMemoryUsage.Set(float64(memUsage))
    
    return nil
}
```

### Analytics Engine

**Analytics Manager**
```go
// pkg/analytics/manager.go
package analytics

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    "github.com/google/uuid"
)

type AnalyticsManager struct {
    storage      AnalyticsStorage
    processor    *StreamProcessor
    calculator   *MetricsCalculator
    config       *Config
    logger       Logger
}

type Config struct {
    Enabled             bool          `yaml:"enabled"`
    ProcessingInterval  time.Duration `yaml:"processing_interval"`
    RetentionDays       int           `yaml:"retention_days"`
    
    // Stream processing
    StreamEnabled       bool          `yaml:"stream_enabled"`
    StreamBuffer        int           `yaml:"stream_buffer"`
    
    // Batch processing
    BatchSize           int           `yaml:"batch_size"`
    BatchInterval       time.Duration `yaml:"batch_interval"`
    
    // Aggregations
    Aggregations        []Aggregation `yaml:"aggregations"`
}

type Aggregation struct {
    Name        string        `yaml:"name"`
    Source      string        `yaml:"source"`
    Metrics     []string      `yaml:"metrics"`
    GroupBy     []string      `yaml:"group_by"`
    Interval    time.Duration `yaml:"interval"`
    Retention   time.Duration `yaml:"retention"`
}

type AnalyticsEvent struct {
    ID          uuid.UUID              `json:"id"`
    Type        string                 `json:"type"`
    UserID      *uuid.UUID             `json:"user_id,omitempty"`
    SessionID   *uuid.UUID             `json:"session_id,omitempty"`
    MissionID   *uuid.UUID             `json:"mission_id,omitempty"`
    Properties  map[string]interface{} `json:"properties"`
    Timestamp   time.Time              `json:"timestamp"`
    ProcessedAt *time.Time             `json:"processed_at,omitempty"`
}

type MetricSnapshot struct {
    Name        string                 `json:"name"`
    Value       float64                `json:"value"`
    Labels      map[string]string      `json:"labels"`
    Timestamp   time.Time              `json:"timestamp"`
    Aggregation string                 `json:"aggregation"`
    Interval    time.Duration          `json:"interval"`
}

type TacticalSituation struct {
    Timestamp       time.Time              `json:"timestamp"`
    ActiveUnits     int                    `json:"active_units"`
    UnitPositions   []UnitPosition         `json:"unit_positions"`
    Communications  CommunicationMetrics   `json:"communications"`
    MissionStatus   MissionStatusSummary   `json:"mission_status"`
    ThreatLevel     string                 `json:"threat_level"`
    Weather         WeatherConditions      `json:"weather,omitempty"`
}

type UnitPosition struct {
    UnitID      string    `json:"unit_id"`
    Callsign    string    `json:"callsign"`
    Type        string    `json:"type"`
    Position    Position  `json:"position"`
    Status      string    `json:"status"`
    LastUpdate  time.Time `json:"last_update"`
}

type CommunicationMetrics struct {
    ActiveChannels    int     `json:"active_channels"`
    MessagesPerMinute float64 `json:"messages_per_minute"`
    NetworkHealth     string  `json:"network_health"`
    Connectivity      float64 `json:"connectivity_percent"`
}

func NewAnalyticsManager(storage AnalyticsStorage, config *Config, logger Logger) *AnalyticsManager {
    return &AnalyticsManager{
        storage:    storage,
        processor:  NewStreamProcessor(config.StreamBuffer, logger),
        calculator: NewMetricsCalculator(config, logger),
        config:     config,
        logger:     logger,
    }
}

func (am *AnalyticsManager) Start(ctx context.Context) error {
    if !am.config.Enabled {
        am.logger.Info("Analytics disabled")
        return nil
    }
    
    // Start stream processor
    if am.config.StreamEnabled {
        go am.processor.Start(ctx)
    }
    
    // Start batch processing
    go am.batchProcessor(ctx)
    
    // Start aggregation calculations
    go am.runAggregations(ctx)
    
    am.logger.Info("Analytics manager started")
    return nil
}

func (am *AnalyticsManager) RecordEvent(ctx context.Context, event *AnalyticsEvent) error {
    event.ID = uuid.New()
    event.Timestamp = time.Now()
    
    // Store event
    if err := am.storage.StoreEvent(ctx, event); err != nil {
        return fmt.Errorf("failed to store analytics event: %w", err)
    }
    
    // Send to stream processor if enabled
    if am.config.StreamEnabled {
        am.processor.Process(event)
    }
    
    return nil
}

func (am *AnalyticsManager) GetTacticalSituation(ctx context.Context) (*TacticalSituation, error) {
    situation := &TacticalSituation{
        Timestamp: time.Now(),
    }
    
    // Get active units count
    activeUnits, err := am.storage.GetActiveUnitsCount(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get active units count: %w", err)
    }
    situation.ActiveUnits = activeUnits
    
    // Get unit positions
    positions, err := am.storage.GetCurrentUnitPositions(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get unit positions: %w", err)
    }
    situation.UnitPositions = positions
    
    // Get communication metrics
    commMetrics, err := am.calculateCommunicationMetrics(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate communication metrics: %w", err)
    }
    situation.Communications = commMetrics
    
    // Get mission status
    missionStatus, err := am.getMissionStatusSummary(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get mission status: %w", err)
    }
    situation.MissionStatus = missionStatus
    
    // Calculate threat level
    situation.ThreatLevel = am.calculateThreatLevel(situation)
    
    return situation, nil
}

func (am *AnalyticsManager) GenerateUsageReport(ctx context.Context, startTime, endTime time.Time) (*UsageReport, error) {
    report := &UsageReport{
        Period: Period{
            Start: startTime,
            End:   endTime,
        },
        GeneratedAt: time.Now(),
    }
    
    // User activity metrics
    userMetrics, err := am.storage.GetUserActivityMetrics(ctx, startTime, endTime)
    if err != nil {
        return nil, fmt.Errorf("failed to get user activity metrics: %w", err)
    }
    report.UserActivity = userMetrics
    
    // System utilization metrics
    systemMetrics, err := am.storage.GetSystemUtilizationMetrics(ctx, startTime, endTime)
    if err != nil {
        return nil, fmt.Errorf("failed to get system utilization metrics: %w", err)
    }
    report.SystemUtilization = systemMetrics
    
    // Mission effectiveness metrics
    missionMetrics, err := am.storage.GetMissionEffectivenessMetrics(ctx, startTime, endTime)
    if err != nil {
        return nil, fmt.Errorf("failed to get mission effectiveness metrics: %w", err)
    }
    report.MissionEffectiveness = missionMetrics
    
    // Resource optimization recommendations
    recommendations, err := am.generateRecommendations(ctx, report)
    if err != nil {
        am.logger.Warn("Failed to generate recommendations", "error", err)
    } else {
        report.Recommendations = recommendations
    }
    
    return report, nil
}

func (am *AnalyticsManager) batchProcessor(ctx context.Context) {
    ticker := time.NewTicker(am.config.BatchInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := am.processBatch(ctx); err != nil {
                am.logger.Error("Failed to process batch", "error", err)
            }
        }
    }
}

func (am *AnalyticsManager) processBatch(ctx context.Context) error {
    // Get unprocessed events
    events, err := am.storage.GetUnprocessedEvents(ctx, am.config.BatchSize)
    if err != nil {
        return fmt.Errorf("failed to get unprocessed events: %w", err)
    }
    
    if len(events) == 0 {
        return nil
    }
    
    am.logger.Debug("Processing batch", "count", len(events))
    
    // Process each event
    for _, event := range events {
        if err := am.processEvent(ctx, event); err != nil {
            am.logger.Error("Failed to process event", "event_id", event.ID, "error", err)
            continue
        }
        
        // Mark as processed
        now := time.Now()
        event.ProcessedAt = &now
        if err := am.storage.UpdateEvent(ctx, event); err != nil {
            am.logger.Error("Failed to update event", "event_id", event.ID, "error", err)
        }
    }
    
    return nil
}

func (am *AnalyticsManager) processEvent(ctx context.Context, event *AnalyticsEvent) error {
    switch event.Type {
    case "user_login":
        return am.processUserLogin(ctx, event)
    case "user_action":
        return am.processUserAction(ctx, event)
    case "message_sent":
        return am.processMessageSent(ctx, event)
    case "mission_update":
        return am.processMissionUpdate(ctx, event)
    case "position_update":
        return am.processPositionUpdate(ctx, event)
    default:
        am.logger.Debug("Unknown event type", "type", event.Type)
        return nil
    }
}

type UsageReport struct {
    Period                Period                    `json:"period"`
    GeneratedAt           time.Time                 `json:"generated_at"`
    UserActivity          UserActivityMetrics       `json:"user_activity"`
    SystemUtilization     SystemUtilizationMetrics  `json:"system_utilization"`
    MissionEffectiveness  MissionEffectivenessMetrics `json:"mission_effectiveness"`
    Recommendations       []Recommendation          `json:"recommendations"`
}

type UserActivityMetrics struct {
    TotalUsers        int                 `json:"total_users"`
    ActiveUsers       int                 `json:"active_users"`
    AverageSessionTime time.Duration      `json:"average_session_time"`
    TopActions        []ActionCount       `json:"top_actions"`
    UsersByRole       map[string]int      `json:"users_by_role"`
    LoginsByHour      []HourlyCount       `json:"logins_by_hour"`
}

type SystemUtilizationMetrics struct {
    AverageCPU        float64             `json:"average_cpu"`
    AverageMemory     float64             `json:"average_memory"`
    PeakConnections   int                 `json:"peak_connections"`
    MessageThroughput float64             `json:"message_throughput"`
    DatabasePerformance DatabaseMetrics   `json:"database_performance"`
}
```

### Dashboard Engine

**Dashboard Manager**
```go
// pkg/dashboard/manager.go
package dashboard

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/google/uuid"
)

type DashboardManager struct {
    storage         DashboardStorage
    widgetManager   *WidgetManager
    dataProvider    *DataProvider
    config          *Config
    logger          Logger
}

type Config struct {
    RefreshInterval    time.Duration    `yaml:"refresh_interval"`
    MaxWidgets         int              `yaml:"max_widgets"`
    CacheTimeout       time.Duration    `yaml:"cache_timeout"`
    DefaultDashboards  []DashboardTemplate `yaml:"default_dashboards"`
}

type Dashboard struct {
    ID          uuid.UUID       `json:"id"`
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Type        string          `json:"type"` // tactical, operational, strategic, system
    Layout      DashboardLayout `json:"layout"`
    Widgets     []Widget        `json:"widgets"`
    Permissions []Permission    `json:"permissions"`
    CreatedBy   uuid.UUID       `json:"created_by"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
}

type DashboardLayout struct {
    Type        string      `json:"type"` // grid, flexible
    Columns     int         `json:"columns"`
    RowHeight   int         `json:"row_height"`
    Margin      [2]int      `json:"margin"`
    Padding     [2]int      `json:"padding"`
}

type Widget struct {
    ID          uuid.UUID       `json:"id"`
    Type        string          `json:"type"`
    Title       string          `json:"title"`
    Position    WidgetPosition  `json:"position"`
    Size        WidgetSize      `json:"size"`
    Config      WidgetConfig    `json:"config"`
    DataSource  DataSource      `json:"data_source"`
    RefreshRate time.Duration   `json:"refresh_rate"`
    LastUpdated time.Time       `json:"last_updated"`
}

type WidgetPosition struct {
    X int `json:"x"`
    Y int `json:"y"`
}

type WidgetSize struct {
    Width  int `json:"width"`
    Height int `json:"height"`
}

type WidgetConfig struct {
    ChartType    string                 `json:"chart_type,omitempty"`
    Colors       []string               `json:"colors,omitempty"`
    Axes         map[string]AxisConfig  `json:"axes,omitempty"`
    Filters      []Filter               `json:"filters,omitempty"`
    Aggregation  string                 `json:"aggregation,omitempty"`
    TimeRange    string                 `json:"time_range,omitempty"`
    Thresholds   []Threshold            `json:"thresholds,omitempty"`
    DisplayMode  string                 `json:"display_mode,omitempty"`
    Properties   map[string]interface{} `json:"properties,omitempty"`
}

type DataSource struct {
    Type       string                 `json:"type"`
    Query      string                 `json:"query"`
    Parameters map[string]interface{} `json:"parameters"`
    Metrics    []string               `json:"metrics"`
    GroupBy    []string               `json:"group_by"`
    TimeField  string                 `json:"time_field"`
}

// Predefined dashboard templates
var TacticalDashboardTemplate = DashboardTemplate{
    Name: "Tactical Operations",
    Type: "tactical",
    Widgets: []WidgetTemplate{
        {
            Type:  "map",
            Title: "Unit Positions",
            Position: WidgetPosition{X: 0, Y: 0},
            Size: WidgetSize{Width: 8, Height: 6},
            DataSource: DataSource{
                Type:  "positions",
                Query: "SELECT * FROM unit_positions WHERE last_update > ?",
                Parameters: map[string]interface{}{
                    "time_threshold": "5m",
                },
            },
        },
        {
            Type:  "gauge",
            Title: "Network Health",
            Position: WidgetPosition{X: 8, Y: 0},
            Size: WidgetSize{Width: 4, Height: 3},
            DataSource: DataSource{
                Type: "metrics",
                Metrics: []string{"network_connectivity_percent"},
            },
        },
        {
            Type:  "timeline",
            Title: "Recent Events",
            Position: WidgetPosition{X: 8, Y: 3},
            Size: WidgetSize{Width: 4, Height: 3},
            DataSource: DataSource{
                Type:  "events",
                Query: "SELECT * FROM events ORDER BY timestamp DESC LIMIT 50",
            },
        },
        {
            Type:  "bar_chart",
            Title: "Active Units by Type",
            Position: WidgetPosition{X: 0, Y: 6},
            Size: WidgetSize{Width: 6, Height: 4},
            DataSource: DataSource{
                Type: "analytics",
                Query: "SELECT unit_type, COUNT(*) FROM units WHERE status = 'active' GROUP BY unit_type",
            },
        },
    },
}

func NewDashboardManager(storage DashboardStorage, config *Config, logger Logger) *DashboardManager {
    return &DashboardManager{
        storage:       storage,
        widgetManager: NewWidgetManager(logger),
        dataProvider:  NewDataProvider(logger),
        config:        config,
        logger:        logger,
    }
}

func (dm *DashboardManager) CreateDashboard(ctx context.Context, userID uuid.UUID, template DashboardTemplate) (*Dashboard, error) {
    dashboard := &Dashboard{
        ID:          uuid.New(),
        Name:        template.Name,
        Description: template.Description,
        Type:        template.Type,
        Layout:      template.Layout,
        CreatedBy:   userID,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    // Create widgets from template
    widgets := make([]Widget, 0, len(template.Widgets))
    for _, widgetTemplate := range template.Widgets {
        widget := Widget{
            ID:          uuid.New(),
            Type:        widgetTemplate.Type,
            Title:       widgetTemplate.Title,
            Position:    widgetTemplate.Position,
            Size:        widgetTemplate.Size,
            Config:      widgetTemplate.Config,
            DataSource:  widgetTemplate.DataSource,
            RefreshRate: widgetTemplate.RefreshRate,
            LastUpdated: time.Now(),
        }
        widgets = append(widgets, widget)
    }
    dashboard.Widgets = widgets
    
    // Store dashboard
    if err := dm.storage.CreateDashboard(ctx, dashboard); err != nil {
        return nil, fmt.Errorf("failed to create dashboard: %w", err)
    }
    
    dm.logger.Info("Dashboard created", "dashboard_id", dashboard.ID, "name", dashboard.Name, "user_id", userID)
    
    return dashboard, nil
}

func (dm *DashboardManager) GetDashboardData(ctx context.Context, dashboardID uuid.UUID) (*DashboardData, error) {
    // Get dashboard configuration
    dashboard, err := dm.storage.GetDashboard(ctx, dashboardID)
    if err != nil {
        return nil, fmt.Errorf("failed to get dashboard: %w", err)
    }
    
    data := &DashboardData{
        Dashboard: dashboard,
        Data:      make(map[uuid.UUID]interface{}),
        UpdatedAt: time.Now(),
    }
    
    // Get data for each widget
    for _, widget := range dashboard.Widgets {
        widgetData, err := dm.getWidgetData(ctx, widget)
        if err != nil {
            dm.logger.Error("Failed to get widget data", "widget_id", widget.ID, "error", err)
            continue
        }
        data.Data[widget.ID] = widgetData
    }
    
    return data, nil
}

func (dm *DashboardManager) getWidgetData(ctx context.Context, widget Widget) (interface{}, error) {
    switch widget.DataSource.Type {
    case "metrics":
        return dm.dataProvider.GetMetricsData(ctx, widget.DataSource)
    case "analytics":
        return dm.dataProvider.GetAnalyticsData(ctx, widget.DataSource)
    case "positions":
        return dm.dataProvider.GetPositionData(ctx, widget.DataSource)
    case "events":
        return dm.dataProvider.GetEventData(ctx, widget.DataSource)
    case "missions":
        return dm.dataProvider.GetMissionData(ctx, widget.DataSource)
    default:
        return nil, fmt.Errorf("unknown data source type: %s", widget.DataSource.Type)
    }
}

type DashboardData struct {
    Dashboard *Dashboard                  `json:"dashboard"`
    Data      map[uuid.UUID]interface{}   `json:"data"`
    UpdatedAt time.Time                   `json:"updated_at"`
}

type WidgetData struct {
    Type       string      `json:"type"`
    Data       interface{} `json:"data"`
    UpdatedAt  time.Time   `json:"updated_at"`
    Error      string      `json:"error,omitempty"`
}
```

### Alerting System

**Alert Manager**
```go
// pkg/alerting/manager.go
package alerting

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/uuid"
)

type AlertManager struct {
    storage         AlertStorage
    ruleEngine      *RuleEngine
    notifier        *NotificationManager
    escalator       *EscalationManager
    config          *Config
    logger          Logger
}

type Config struct {
    Enabled            bool          `yaml:"enabled"`
    EvaluationInterval time.Duration `yaml:"evaluation_interval"`
    DefaultSeverity    string        `yaml:"default_severity"`
    RetentionDays      int           `yaml:"retention_days"`
    
    // Notification settings
    NotificationChannels []NotificationChannel `yaml:"notification_channels"`
    EscalationPolicies   []EscalationPolicy    `yaml:"escalation_policies"`
    
    // Anti-spam settings
    RateLimits         map[string]RateLimit  `yaml:"rate_limits"`
    DeduplicationWindow time.Duration        `yaml:"deduplication_window"`
}

type AlertRule struct {
    ID          uuid.UUID   `json:"id"`
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Query       string      `json:"query"`
    Condition   Condition   `json:"condition"`
    Severity    string      `json:"severity"`
    Category    string      `json:"category"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    
    // Evaluation
    EvaluateEvery  time.Duration `json:"evaluate_every"`
    EvaluateFor    time.Duration `json:"evaluate_for"`
    
    // State
    State       string    `json:"state"`
    LastEval    time.Time `json:"last_eval"`
    
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Condition struct {
    Operator    string  `json:"operator"` // gt, lt, eq, ne
    Threshold   float64 `json:"threshold"`
    Aggregation string  `json:"aggregation"` // avg, sum, count, min, max
}

type Alert struct {
    ID          uuid.UUID         `json:"id"`
    RuleID      uuid.UUID         `json:"rule_id"`
    RuleName    string            `json:"rule_name"`
    Severity    string            `json:"severity"`
    Category    string            `json:"category"`
    Message     string            `json:"message"`
    Description string            `json:"description"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    
    // Values
    CurrentValue   float64                `json:"current_value"`
    ThresholdValue float64                `json:"threshold_value"`
    QueryResult    map[string]interface{} `json:"query_result"`
    
    // State
    State       AlertState `json:"state"`
    FiredAt     time.Time  `json:"fired_at"`
    ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
    AckedAt     *time.Time `json:"acked_at,omitempty"`
    AckedBy     *uuid.UUID `json:"acked_by,omitempty"`
    
    // Notifications
    NotificationsSent []NotificationRecord `json:"notifications_sent"`
    Escalated         bool                 `json:"escalated"`
    EscalationLevel   int                  `json:"escalation_level"`
    
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type AlertState string
const (
    AlertStatePending   AlertState = "pending"
    AlertStateFiring    AlertState = "firing"
    AlertStateResolved  AlertState = "resolved"
    AlertStateSuppressed AlertState = "suppressed"
)

type NotificationRecord struct {
    Channel   string    `json:"channel"`
    SentAt    time.Time `json:"sent_at"`
    Status    string    `json:"status"`
    Error     string    `json:"error,omitempty"`
}

// Predefined alert rules
var SystemAlertRules = []AlertRule{
    {
        Name:        "High CPU Usage",
        Description: "System CPU usage is above 80%",
        Query:       "avg(gotak_system_cpu_usage_percent)",
        Condition: Condition{
            Operator:    "gt",
            Threshold:   80,
            Aggregation: "avg",
        },
        Severity:      "warning",
        Category:      "system",
        EvaluateEvery: time.Minute,
        EvaluateFor:   time.Minute * 5,
    },
    {
        Name:        "High Memory Usage",
        Description: "System memory usage is above 90%",
        Query:       "avg(gotak_system_memory_usage_percent)",
        Condition: Condition{
            Operator:    "gt",
            Threshold:   90,
            Aggregation: "avg",
        },
        Severity:      "critical",
        Category:      "system",
        EvaluateEvery: time.Minute,
        EvaluateFor:   time.Minute * 3,
    },
    {
        Name:        "Database Connection Pool Full",
        Description: "Database connection pool is at capacity",
        Query:       "avg(gotak_database_connections{state=\"active\"}) / avg(gotak_database_connections{state=\"max\"})",
        Condition: Condition{
            Operator:    "gt",
            Threshold:   0.95,
            Aggregation: "avg",
        },
        Severity:      "critical",
        Category:      "database",
        EvaluateEvery: time.Minute,
        EvaluateFor:   time.Minute * 2,
    },
    {
        Name:        "Federation Connection Lost",
        Description: "Federation connection has been lost",
        Query:       "sum(gotak_federation_connections{status=\"connected\"})",
        Condition: Condition{
            Operator:    "lt",
            Threshold:   1,
            Aggregation: "sum",
        },
        Severity:      "warning",
        Category:      "federation",
        EvaluateEvery: time.Minute,
        EvaluateFor:   time.Minute * 5,
    },
    {
        Name:        "No User Activity",
        Description: "No user activity detected for extended period",
        Query:       "sum(rate(gotak_user_actions_total[5m]))",
        Condition: Condition{
            Operator:    "lt",
            Threshold:   0.1,
            Aggregation: "sum",
        },
        Severity:      "info",
        Category:      "operational",
        EvaluateEvery: time.Minute * 5,
        EvaluateFor:   time.Minute * 15,
    },
}

func NewAlertManager(storage AlertStorage, config *Config, logger Logger) *AlertManager {
    return &AlertManager{
        storage:    storage,
        ruleEngine: NewRuleEngine(logger),
        notifier:   NewNotificationManager(config.NotificationChannels, logger),
        escalator:  NewEscalationManager(config.EscalationPolicies, logger),
        config:     config,
        logger:     logger,
    }
}

func (am *AlertManager) Start(ctx context.Context) error {
    if !am.config.Enabled {
        am.logger.Info("Alerting disabled")
        return nil
    }
    
    // Load alert rules
    if err := am.loadDefaultRules(ctx); err != nil {
        return fmt.Errorf("failed to load default rules: %w", err)
    }
    
    // Start evaluation loop
    go am.evaluationLoop(ctx)
    
    // Start escalation processor
    go am.escalationLoop(ctx)
    
    // Start cleanup routine
    go am.cleanupLoop(ctx)
    
    am.logger.Info("Alert manager started")
    return nil
}

func (am *AlertManager) evaluationLoop(ctx context.Context) {
    ticker := time.NewTicker(am.config.EvaluationInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := am.evaluateRules(ctx); err != nil {
                am.logger.Error("Failed to evaluate rules", "error", err)
            }
        }
    }
}

func (am *AlertManager) evaluateRules(ctx context.Context) error {
    rules, err := am.storage.GetActiveRules(ctx)
    if err != nil {
        return fmt.Errorf("failed to get active rules: %w", err)
    }
    
    for _, rule := range rules {
        go func(r AlertRule) {
            if err := am.evaluateRule(ctx, &r); err != nil {
                am.logger.Error("Failed to evaluate rule", "rule_id", r.ID, "error", err)
            }
        }(rule)
    }
    
    return nil
}

func (am *AlertManager) evaluateRule(ctx context.Context, rule *AlertRule) error {
    // Execute query
    result, err := am.ruleEngine.ExecuteQuery(ctx, rule.Query)
    if err != nil {
        return fmt.Errorf("failed to execute query: %w", err)
    }
    
    // Evaluate condition
    shouldFire := am.ruleEngine.EvaluateCondition(result, rule.Condition)
    
    // Get existing alert for this rule
    existingAlert, err := am.storage.GetActiveAlertByRule(ctx, rule.ID)
    if err != nil && err != ErrAlertNotFound {
        return fmt.Errorf("failed to get existing alert: %w", err)
    }
    
    if shouldFire {
        if existingAlert == nil {
            // Create new alert
            alert := &Alert{
                ID:             uuid.New(),
                RuleID:         rule.ID,
                RuleName:       rule.Name,
                Severity:       rule.Severity,
                Category:       rule.Category,
                Message:        am.buildAlertMessage(rule, result),
                Description:    rule.Description,
                Labels:         rule.Labels,
                Annotations:    rule.Annotations,
                CurrentValue:   result.Value,
                ThresholdValue: rule.Condition.Threshold,
                QueryResult:    result.Data,
                State:          AlertStatePending,
                FiredAt:        time.Now(),
                CreatedAt:      time.Now(),
                UpdatedAt:      time.Now(),
            }
            
            if err := am.storage.CreateAlert(ctx, alert); err != nil {
                return fmt.Errorf("failed to create alert: %w", err)
            }
            
            // Check if alert should be fired immediately or wait for EvaluateFor duration
            if rule.EvaluateFor == 0 {
                alert.State = AlertStateFiring
                if err := am.fireAlert(ctx, alert); err != nil {
                    am.logger.Error("Failed to fire alert", "alert_id", alert.ID, "error", err)
                }
            }
        } else {
            // Update existing alert
            existingAlert.CurrentValue = result.Value
            existingAlert.QueryResult = result.Data
            existingAlert.UpdatedAt = time.Now()
            
            // Check if pending alert should be fired
            if existingAlert.State == AlertStatePending &&
               time.Since(existingAlert.FiredAt) >= rule.EvaluateFor {
                existingAlert.State = AlertStateFiring
                if err := am.fireAlert(ctx, existingAlert); err != nil {
                    am.logger.Error("Failed to fire alert", "alert_id", existingAlert.ID, "error", err)
                }
            }
            
            if err := am.storage.UpdateAlert(ctx, existingAlert); err != nil {
                return fmt.Errorf("failed to update alert: %w", err)
            }
        }
    } else {
        // Condition not met - resolve existing alert if any
        if existingAlert != nil && existingAlert.State != AlertStateResolved {
            existingAlert.State = AlertStateResolved
            now := time.Now()
            existingAlert.ResolvedAt = &now
            existingAlert.UpdatedAt = now
            
            if err := am.storage.UpdateAlert(ctx, existingAlert); err != nil {
                return fmt.Errorf("failed to resolve alert: %w", err)
            }
            
            // Send resolution notification
            if err := am.sendResolutionNotification(ctx, existingAlert); err != nil {
                am.logger.Error("Failed to send resolution notification", "alert_id", existingAlert.ID, "error", err)
            }
        }
    }
    
    // Update rule last evaluation time
    rule.LastEval = time.Now()
    if err := am.storage.UpdateRule(ctx, rule); err != nil {
        am.logger.Error("Failed to update rule", "rule_id", rule.ID, "error", err)
    }
    
    return nil
}

func (am *AlertManager) fireAlert(ctx context.Context, alert *Alert) error {
    am.logger.Info("Firing alert", "alert_id", alert.ID, "rule", alert.RuleName, "severity", alert.Severity)
    
    // Send notifications
    if err := am.notifier.SendAlert(ctx, alert); err != nil {
        return fmt.Errorf("failed to send alert notifications: %w", err)
    }
    
    // Schedule escalation if configured
    if err := am.escalator.ScheduleEscalation(ctx, alert); err != nil {
        am.logger.Error("Failed to schedule escalation", "alert_id", alert.ID, "error", err)
    }
    
    return nil
}
```

## Database Schema

```sql
-- Analytics events
CREATE TABLE analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id UUID,
    mission_id UUID REFERENCES missions(id) ON DELETE SET NULL,
    properties JSONB,
    timestamp TIMESTAMP NOT NULL,
    processed_at TIMESTAMP,
    
    INDEX(type, timestamp),
    INDEX(user_id, timestamp),
    INDEX(processed_at)
);

-- Metric snapshots
CREATE TABLE metric_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    value DECIMAL(15,6) NOT NULL,
    labels JSONB,
    timestamp TIMESTAMP NOT NULL,
    aggregation VARCHAR(50),
    interval_seconds INTEGER,
    
    INDEX(name, timestamp),
    INDEX(timestamp)
);

-- Dashboards
CREATE TABLE dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    layout JSONB NOT NULL,
    widgets JSONB NOT NULL,
    permissions JSONB,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Alert rules
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    query TEXT NOT NULL,
    condition JSONB NOT NULL,
    severity VARCHAR(20) NOT NULL,
    category VARCHAR(50) NOT NULL,
    labels JSONB,
    annotations JSONB,
    evaluate_every INTERVAL NOT NULL,
    evaluate_for INTERVAL NOT NULL,
    state VARCHAR(20) DEFAULT 'active',
    last_eval TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Alerts
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID REFERENCES alert_rules(id) ON DELETE CASCADE,
    rule_name VARCHAR(255) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    category VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    description TEXT,
    labels JSONB,
    annotations JSONB,
    current_value DECIMAL(15,6),
    threshold_value DECIMAL(15,6),
    query_result JSONB,
    state VARCHAR(20) NOT NULL,
    fired_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    acked_at TIMESTAMP,
    acked_by UUID REFERENCES users(id),
    notifications_sent JSONB,
    escalated BOOLEAN DEFAULT false,
    escalation_level INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Unit positions (for tactical analytics)
CREATE TABLE unit_positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id VARCHAR(255) NOT NULL,
    callsign VARCHAR(255) NOT NULL,
    unit_type VARCHAR(100),
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    altitude DECIMAL(10, 2),
    course DECIMAL(5, 2),
    speed DECIMAL(8, 2),
    status VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    
    INDEX(unit_id, timestamp DESC),
    INDEX(timestamp),
    INDEX(unit_type)
);

-- Performance optimization indexes
CREATE INDEX idx_analytics_events_type_time ON analytics_events(type, timestamp DESC);
CREATE INDEX idx_analytics_events_user_time ON analytics_events(user_id, timestamp DESC);
CREATE INDEX idx_metric_snapshots_name_time ON metric_snapshots(name, timestamp DESC);
CREATE INDEX idx_alerts_rule_state ON alerts(rule_id, state);
CREATE INDEX idx_alerts_severity_time ON alerts(severity, fired_at DESC);

-- Time-series partitioning for analytics_events (PostgreSQL 10+)
CREATE TABLE analytics_events_y2025m01 PARTITION OF analytics_events
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE analytics_events_y2025m02 PARTITION OF analytics_events
FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
```

## API Specifications

### Analytics API
```
GET    /api/v1/analytics/situation           # Current tactical situation
GET    /api/v1/analytics/usage-report        # Usage analytics report
POST   /api/v1/analytics/events              # Record analytics event
GET    /api/v1/analytics/metrics             # Query metrics data
GET    /api/v1/analytics/trends              # Trend analysis
```

### Dashboard API
```
GET    /api/v1/dashboards                    # List user dashboards
POST   /api/v1/dashboards                    # Create dashboard
GET    /api/v1/dashboards/{id}               # Get dashboard
PUT    /api/v1/dashboards/{id}               # Update dashboard
DELETE /api/v1/dashboards/{id}               # Delete dashboard
GET    /api/v1/dashboards/{id}/data          # Get dashboard data
GET    /api/v1/dashboards/templates          # List dashboard templates
```

### Alerting API
```
GET    /api/v1/alerts                        # List alerts
GET    /api/v1/alerts/{id}                   # Get alert details
POST   /api/v1/alerts/{id}/ack               # Acknowledge alert
POST   /api/v1/alerts/{id}/resolve           # Resolve alert
GET    /api/v1/alert-rules                   # List alert rules
POST   /api/v1/alert-rules                   # Create alert rule
PUT    /api/v1/alert-rules/{id}              # Update alert rule
DELETE /api/v1/alert-rules/{id}              # Delete alert rule
```

### Metrics API
```
GET    /metrics                              # Prometheus metrics endpoint
GET    /api/v1/metrics/query                 # Query metrics
GET    /api/v1/metrics/query_range           # Query metrics over time range
GET    /api/v1/metrics/labels                # Get metric labels
GET    /api/v1/metrics/values                # Get label values
```

## Testing Strategy

### Unit Tests
```go
func TestAnalyticsManager_RecordEvent(t *testing.T) {
    manager := setupTestAnalyticsManager()
    
    event := &AnalyticsEvent{
        Type:   "user_login",
        UserID: &uuid.New(),
        Properties: map[string]interface{}{
            "source": "web",
            "method": "password",
        },
    }
    
    err := manager.RecordEvent(context.Background(), event)
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, event.ID)
    assert.False(t, event.Timestamp.IsZero())
}

func TestDashboardManager_CreateDashboard(t *testing.T) {
    manager := setupTestDashboardManager()
    userID := uuid.New()
    
    dashboard, err := manager.CreateDashboard(context.Background(), userID, TacticalDashboardTemplate)
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, dashboard.ID)
    assert.Equal(t, "Tactical Operations", dashboard.Name)
    assert.Len(t, dashboard.Widgets, 4)
}

func TestAlertManager_EvaluateRule(t *testing.T) {
    manager := setupTestAlertManager()
    
    rule := &AlertRule{
        ID:    uuid.New(),
        Name:  "Test Rule",
        Query: "SELECT 95.0 as value",
        Condition: Condition{
            Operator:  "gt",
            Threshold: 90.0,
        },
        Severity:      "warning",
        EvaluateEvery: time.Minute,
        EvaluateFor:   0,
    }
    
    err := manager.evaluateRule(context.Background(), rule)
    assert.NoError(t, err)
    
    // Check that alert was created
    alerts, err := manager.storage.GetAlertsByRule(context.Background(), rule.ID)
    assert.NoError(t, err)
    assert.Len(t, alerts, 1)
    assert.Equal(t, AlertStateFiring, alerts[0].State)
}
```

### Integration Tests
```go
func TestEndToEndMonitoring(t *testing.T) {
    // Start monitoring system
    metrics := startMetricsManager()
    analytics := startAnalyticsManager()
    alerts := startAlertManager()
    
    defer stopAll(metrics, analytics, alerts)
    
    // Generate some test data
    generateTestMetrics()
    generateTestEvents()
    
    // Wait for processing
    time.Sleep(5 * time.Second)
    
    // Verify metrics collected
    cpuMetric := getMetricValue("gotak_system_cpu_usage_percent")
    assert.True(t, cpuMetric > 0)
    
    // Verify events processed
    events := getProcessedEvents("user_login")
    assert.True(t, len(events) > 0)
    
    // Verify alerts triggered if thresholds exceeded
    alerts := getActiveAlerts()
    for _, alert := range alerts {
        assert.True(t, alert.CurrentValue > alert.ThresholdValue)
    }
}
```

## Acceptance Criteria

### Real-Time Monitoring
- [ ] System metrics collected and visualized in real-time
- [ ] Application performance metrics tracking
- [ ] Database performance monitoring
- [ ] Custom metrics registration and collection
- [ ] Prometheus metrics endpoint operational

### Analytics Dashboard
- [ ] Interactive tactical operations dashboard
- [ ] Customizable widget-based dashboards
- [ ] Real-time data refresh and visualization
- [ ] Dashboard templates for common use cases
- [ ] User permissions and sharing capabilities

### Alerting System
- [ ] Rule-based alerting with configurable thresholds
- [ ] Multi-channel notification delivery
- [ ] Alert escalation and acknowledgment
- [ ] Anomaly detection for unusual patterns
- [ ] Alert correlation and deduplication

### Business Intelligence
- [ ] Usage analytics and trend analysis
- [ ] Mission effectiveness reporting
- [ ] Resource utilization optimization
- [ ] Automated report generation and distribution
- [ ] Executive dashboard summaries

### Performance
- [ ] Real-time dashboard updates (< 5 seconds)
- [ ] Analytics processing handles 10,000+ events/minute
- [ ] Alert evaluation completes within 30 seconds
- [ ] Dashboard rendering time < 2 seconds
- [ ] Metrics collection adds < 5% overhead

## Dependencies

### Backend Dependencies
```go
require (
    github.com/prometheus/client_golang v1.17.0      // Metrics collection
    github.com/prometheus/common v0.44.0             // Prometheus utilities
    github.com/influxdata/influxdb-client-go/v2 v2.12.3 // InfluxDB client
    github.com/grafana/grafana-api-golang-client v0.23.0 // Grafana integration
)
```

### Infrastructure Dependencies
- Time-series database (Prometheus, InfluxDB)
- Visualization platform (Grafana)
- Message queue for event streaming (Redis, Kafka)
- Notification services (email, SMS, Slack)

## Definition of Done

### Code Quality
- [ ] All code reviewed and approved
- [ ] Unit tests with 85%+ coverage
- [ ] Integration tests for monitoring components
- [ ] Performance benchmarks meet requirements
- [ ] Load testing completed for analytics pipeline

### Functionality
- [ ] All user stories completed and accepted
- [ ] Real-time monitoring operational
- [ ] Analytics dashboards functional
- [ ] Alerting system active and tested
- [ ] Automated reporting working

### Operations
- [ ] Monitoring system self-monitoring configured
- [ ] Alerting rules tuned to reduce false positives
- [ ] Dashboard templates created for all user roles
- [ ] Performance optimization completed
- [ ] Documentation complete for operations team

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Next Sprint Planning:** [TBD]
