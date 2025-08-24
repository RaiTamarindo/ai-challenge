package com.featurevoting.models

import com.google.gson.annotations.SerializedName

data class Feature(
    @SerializedName("id")
    val id: Int,
    
    @SerializedName("title")
    val title: String,
    
    @SerializedName("description")
    val description: String,
    
    @SerializedName("created_by")
    val createdBy: Int,
    
    @SerializedName("created_by_user")
    val createdByUser: String,
    
    @SerializedName("vote_count")
    val voteCount: Int,
    
    @SerializedName("has_user_voted")
    val hasUserVoted: Boolean,
    
    @SerializedName("created_at")
    val createdAt: String,
    
    @SerializedName("updated_at")
    val updatedAt: String
)

data class CreateFeatureRequest(
    @SerializedName("title")
    val title: String,
    
    @SerializedName("description")
    val description: String
)

data class CreateFeatureResponse(
    @SerializedName("message")
    val message: String
)

data class FeaturesResponse(
    @SerializedName("features")
    val features: List<Feature>,
    
    @SerializedName("total")
    val total: Int,
    
    @SerializedName("page")
    val page: Int,
    
    @SerializedName("per_page")
    val perPage: Int
)

data class VoteResponse(
    @SerializedName("message")
    val message: String,
    
    @SerializedName("feature_id")
    val featureId: Int,
    
    @SerializedName("vote_count")
    val voteCount: Int,
    
    @SerializedName("has_voted")
    val hasVoted: Boolean
)