/**
 * DrawingTools Component
 * Tactical drawing and annotation tools for the map using leaflet-draw
 */

import React, { useEffect, useRef, useState } from 'react';
import { useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet-draw/dist/leaflet.draw.css';
import 'leaflet-draw';

// Drawing tool types
export type DrawingTool = 'select' | 'marker' | 'polyline' | 'polygon' | 'rectangle' | 'circle' | 'measure';

// Drawing style presets
const DRAWING_STYLES = {
  default: {
    color: '#3388ff',
    weight: 3,
    opacity: 0.8,
    fillOpacity: 0.2,
  },
  tactical: {
    color: '#ff0000',
    weight: 2,
    opacity: 0.9,
    fillOpacity: 0.3,
  },
  friendly: {
    color: '#0066cc',
    weight: 2,
    opacity: 0.9,
    fillOpacity: 0.3,
  },
  hostile: {
    color: '#cc0000',
    weight: 2,
    opacity: 0.9,
    fillOpacity: 0.3,
  },
  neutral: {
    color: '#ffaa00',
    weight: 2,
    opacity: 0.9,
    fillOpacity: 0.3,
  },
};

// Drawing annotation interface
export interface DrawingAnnotation {
  id: string;
  type: string;
  layer: L.Layer;
  properties: {
    name?: string;
    description?: string;
    category?: 'tactical' | 'friendly' | 'hostile' | 'neutral' | 'default';
    timestamp: string;
    author?: string;
  };
}

interface DrawingToolsProps {
  activeTool: DrawingTool;
  onToolChange?: (tool: DrawingTool) => void;
  onDrawingCreated?: (annotation: DrawingAnnotation) => void;
  onDrawingEdited?: (annotation: DrawingAnnotation) => void;
  onDrawingDeleted?: (annotationId: string) => void;
  enabled?: boolean;
}

const DrawingTools: React.FC<DrawingToolsProps> = ({
  activeTool,
  onToolChange,
  onDrawingCreated,
  onDrawingEdited,
  onDrawingDeleted,
  enabled = true,
}) => {
  const map = useMap();
  const drawControlRef = useRef<L.Control.Draw | null>(null);
  const drawnItemsRef = useRef<L.FeatureGroup>(new L.FeatureGroup());
  const [annotations, setAnnotations] = useState<Map<string, DrawingAnnotation>>(new Map());

  // Initialize drawing controls
  useEffect(() => {
    if (!map || !enabled) return;

    // Add drawn items layer to map
    map.addLayer(drawnItemsRef.current);

    // Create drawing control
    const drawControl = new L.Control.Draw({
      position: 'topright',
      draw: {
        polyline: {
          shapeOptions: DRAWING_STYLES.default,
          metric: true,
          feet: false,
          nautic: false,
        },
        polygon: {
          shapeOptions: DRAWING_STYLES.default,
          allowIntersection: false,
          drawError: {
            color: '#e1e100',
            message: '<strong>Error:</strong> Shape edges cannot cross!',
          },
        },
        circle: {
          shapeOptions: DRAWING_STYLES.default,
          metric: true,
          feet: false,
          nautic: false,
        },
        rectangle: {
          shapeOptions: DRAWING_STYLES.default,
        },
        marker: {
          icon: new L.Icon({
            iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
            shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
            iconSize: [25, 41],
            iconAnchor: [12, 41],
            popupAnchor: [1, -34],
            shadowSize: [41, 41],
          }),
        },
        circlemarker: false, // Disable circle marker
      },
      edit: {
        featureGroup: drawnItemsRef.current,
        remove: true,
      },
    });

    drawControl.addTo(map);
    drawControlRef.current = drawControl;

    return () => {
      if (drawControlRef.current) {
        map.removeControl(drawControlRef.current);
      }
      map.removeLayer(drawnItemsRef.current);
    };
  }, [map, enabled]);

  // Handle drawing events
  useEffect(() => {
    if (!map) return;

    const handleDrawCreated = (e: any) => {
      const { layer, layerType } = e;
      const id = L.stamp(layer).toString();
      
      // Add layer to drawn items
      drawnItemsRef.current.addLayer(layer);
      
      // Create annotation object
      const annotation: DrawingAnnotation = {
        id,
        type: layerType,
        layer,
        properties: {
          category: 'default',
          timestamp: new Date().toISOString(),
        },
      };

      // Add popup for editing properties
      const popupContent = createAnnotationPopup(annotation);
      layer.bindPopup(popupContent);

      // Store annotation
      setAnnotations(prev => new Map(prev.set(id, annotation)));
      
      if (onDrawingCreated) {
        onDrawingCreated(annotation);
      }
    };

    const handleDrawEdited = (e: any) => {
      const layers = e.layers;
      layers.eachLayer((layer: L.Layer) => {
        const id = L.stamp(layer).toString();
        const annotation = annotations.get(id);
        if (annotation && onDrawingEdited) {
          onDrawingEdited({ ...annotation, layer });
        }
      });
    };

    const handleDrawDeleted = (e: any) => {
      const layers = e.layers;
      layers.eachLayer((layer: L.Layer) => {
        const id = L.stamp(layer).toString();
        setAnnotations(prev => {
          const newAnnotations = new Map(prev);
          newAnnotations.delete(id);
          return newAnnotations;
        });
        
        if (onDrawingDeleted) {
          onDrawingDeleted(id);
        }
      });
    };

    // Attach event listeners
    map.on(L.Draw.Event.CREATED, handleDrawCreated);
    map.on(L.Draw.Event.EDITED, handleDrawEdited);
    map.on(L.Draw.Event.DELETED, handleDrawDeleted);

    return () => {
      map.off(L.Draw.Event.CREATED, handleDrawCreated);
      map.off(L.Draw.Event.EDITED, handleDrawEdited);
      map.off(L.Draw.Event.DELETED, handleDrawDeleted);
    };
  }, [map, annotations, onDrawingCreated, onDrawingEdited, onDrawingDeleted]);

  // Create popup content for annotation editing
  const createAnnotationPopup = (annotation: DrawingAnnotation): HTMLElement => {
    const container = document.createElement('div');
    container.className = 'annotation-popup';
    
    container.innerHTML = `
      <div style="min-width: 200px; padding: 8px;">
        <h4 style="margin: 0 0 8px 0; font-size: 14px; font-weight: bold;">
          ${annotation.type.charAt(0).toUpperCase() + annotation.type.slice(1)} Annotation
        </h4>
        
        <div style="margin-bottom: 8px;">
          <label style="display: block; font-weight: bold; margin-bottom: 4px;">Name:</label>
          <input type="text" class="annotation-name" value="${annotation.properties.name || ''}" 
                 style="width: 100%; padding: 4px; border: 1px solid #ccc; border-radius: 2px;" />
        </div>
        
        <div style="margin-bottom: 8px;">
          <label style="display: block; font-weight: bold; margin-bottom: 4px;">Description:</label>
          <textarea class="annotation-description" rows="2" 
                    style="width: 100%; padding: 4px; border: 1px solid #ccc; border-radius: 2px; resize: vertical;">${annotation.properties.description || ''}</textarea>
        </div>
        
        <div style="margin-bottom: 8px;">
          <label style="display: block; font-weight: bold; margin-bottom: 4px;">Category:</label>
          <select class="annotation-category" style="width: 100%; padding: 4px; border: 1px solid #ccc; border-radius: 2px;">
            <option value="default" ${annotation.properties.category === 'default' ? 'selected' : ''}>Default</option>
            <option value="tactical" ${annotation.properties.category === 'tactical' ? 'selected' : ''}>Tactical</option>
            <option value="friendly" ${annotation.properties.category === 'friendly' ? 'selected' : ''}>Friendly</option>
            <option value="hostile" ${annotation.properties.category === 'hostile' ? 'selected' : ''}>Hostile</option>
            <option value="neutral" ${annotation.properties.category === 'neutral' ? 'selected' : ''}>Neutral</option>
          </select>
        </div>
        
        <div style="text-align: center;">
          <button class="save-annotation" style="padding: 6px 12px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; margin-right: 4px;">
            Save
          </button>
          <button class="delete-annotation" style="padding: 6px 12px; background: #dc3545; color: white; border: none; border-radius: 4px; cursor: pointer;">
            Delete
          </button>
        </div>
        
        <div style="margin-top: 8px; font-size: 11px; color: #666;">
          Created: ${new Date(annotation.properties.timestamp).toLocaleString()}
        </div>
      </div>
    `;

    // Add event listeners
    const saveBtn = container.querySelector('.save-annotation') as HTMLButtonElement;
    const deleteBtn = container.querySelector('.delete-annotation') as HTMLButtonElement;
    const nameInput = container.querySelector('.annotation-name') as HTMLInputElement;
    const descInput = container.querySelector('.annotation-description') as HTMLTextAreaElement;
    const categorySelect = container.querySelector('.annotation-category') as HTMLSelectElement;

    if (saveBtn) {
      saveBtn.addEventListener('click', () => {
        // Update annotation properties
        annotation.properties.name = nameInput.value;
        annotation.properties.description = descInput.value;
        annotation.properties.category = categorySelect.value as any;

        // Update layer style based on category
        const style = DRAWING_STYLES[annotation.properties.category] || DRAWING_STYLES.default;
        if ('setStyle' in annotation.layer) {
          (annotation.layer as any).setStyle(style);
        }

        // Update stored annotation
        setAnnotations(prev => new Map(prev.set(annotation.id, annotation)));
        
        // Close popup
        annotation.layer.closePopup();

        if (onDrawingEdited) {
          onDrawingEdited(annotation);
        }
      });
    }

    if (deleteBtn) {
      deleteBtn.addEventListener('click', () => {
        // Remove from drawn items
        drawnItemsRef.current.removeLayer(annotation.layer);
        
        // Remove from annotations
        setAnnotations(prev => {
          const newAnnotations = new Map(prev);
          newAnnotations.delete(annotation.id);
          return newAnnotations;
        });

        if (onDrawingDeleted) {
          onDrawingDeleted(annotation.id);
        }
      });
    }

    return container;
  };

  // Measurement tool functionality
  useEffect(() => {
    if (!map || activeTool !== 'measure') return;

    let measureControl: any = null;

    // Create measurement control
    const createMeasureControl = () => {
      measureControl = new L.Control.Draw({
        position: 'topright',
        draw: {
          polyline: {
            shapeOptions: {
              color: '#ff7800',
              weight: 3,
              opacity: 0.8,
            },
            metric: true,
            feet: false,
            nautic: false,
            showLength: true,
          },
          polygon: false,
          circle: false,
          rectangle: false,
          marker: false,
          circlemarker: false,
        },
        edit: false,
      });

      measureControl.addTo(map);
    };

    createMeasureControl();

    return () => {
      if (measureControl) {
        map.removeControl(measureControl);
      }
    };
  }, [map, activeTool]);

  // Export annotations as GeoJSON
  const exportAnnotations = () => {
    const features = Array.from(annotations.values()).map(annotation => {
      const layer = annotation.layer;
      let geometry: any = null;

      if (layer instanceof L.Marker) {
        const latlng = layer.getLatLng();
        geometry = {
          type: 'Point',
          coordinates: [latlng.lng, latlng.lat],
        };
      } else if (layer instanceof L.Polyline) {
        const latlngs = layer.getLatLngs() as L.LatLng[];
        geometry = {
          type: 'LineString',
          coordinates: latlngs.map(ll => [ll.lng, ll.lat]),
        };
      } else if (layer instanceof L.Polygon) {
        const latlngs = layer.getLatLngs()[0] as L.LatLng[];
        geometry = {
          type: 'Polygon',
          coordinates: [latlngs.map(ll => [ll.lng, ll.lat])],
        };
      }

      return {
        type: 'Feature',
        properties: annotation.properties,
        geometry,
      };
    });

    return {
      type: 'FeatureCollection',
      features,
    };
  };

  return null; // This component doesn't render anything visually
};

export default DrawingTools;
